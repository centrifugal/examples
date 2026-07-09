// End-to-end conformance tester: drives Centrifugo THROUGH the proxy over every
// bidirectional transport and asserts real delivery (not just "connected").
//
// For each (scheme, transport) it:
//   1. connects through the proxy,
//   2. subscribes to a fresh channel,
//   3. publishes via the server HTTP API and asserts the publication arrives
//      within a bounded latency (this is what catches a BUFFERING proxy -
//      "connected OK" proves nothing, a buffering proxy only flushes at close),
//   4. waits past a ping interval and publishes again, asserting the connection
//      stayed alive through an idle window (this catches too-low proxy timeouts).
//
// A separate MODE=crossnode run pins the stream to node-1 and the /emulation
// POST to node-2 to deterministically prove cross-node emulation routing.

import { Centrifuge } from 'centrifuge';
import WebSocket from 'ws';
import { EventSource } from 'eventsource';

const {
  PROXY = 'proxy',
  PROXY_HOST = 'proxy',
  HTTP_PORT = '8080',
  HTTPS_PORT = '8443',
  API_URL = 'http://centrifugo-1:8000',
  API_KEY = '',
  NODE1_URL = 'http://centrifugo-1:8000',
  NODE2_URL = 'http://centrifugo-2:8000',
  MODE = 'matrix',
} = process.env;

const CONNECT_TIMEOUT_MS = 20000;
const RECEIVE_TIMEOUT_MS = 8000;
const LONGEVITY_WAIT_MS = 5000; // > ping_interval (3s) so we cross an idle window.
// Idle-hold mode keeps a connection open with ONLY Centrifugo pings flowing for
// this long, to catch a proxy read/idle timeout that would silently drop it.
// 65s crosses the common 60s default idle timeout (nginx, AWS ALB, ...).
const IDLE_HOLD_MS = parseInt(process.env.IDLE_HOLD_MS || '65000', 10);

const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
let seq = 0;
const uniq = () => `${Date.now().toString(36)}-${seq++}`;

function deferred() {
  let resolve, reject;
  const promise = new Promise((res, rej) => { resolve = res; reject = rej; });
  return { promise, resolve, reject };
}

function withTimeout(promise, ms, label) {
  return Promise.race([
    promise,
    sleep(ms).then(() => { throw new Error(`timeout after ${ms}ms: ${label}`); }),
  ]);
}

async function apiPublish(channel, data) {
  const res = await fetch(`${API_URL}/api/publish`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', 'X-API-Key': API_KEY },
    body: JSON.stringify({ channel, data }),
  });
  if (!res.ok) throw new Error(`publish HTTP ${res.status}`);
  const body = await res.json();
  if (body.error) throw new Error(`publish API error: ${JSON.stringify(body.error)}`);
}

async function apiInfo() {
  const res = await fetch(`${API_URL}/api/info`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', 'X-API-Key': API_KEY },
    body: '{}',
  });
  if (!res.ok) throw new Error(`info HTTP ${res.status}`);
  const body = await res.json();
  return (body.result && body.result.nodes) || [];
}

async function waitReady() {
  const deadline = Date.now() + 60000;
  for (const url of [NODE1_URL, NODE2_URL]) {
    while (true) {
      try {
        const res = await fetch(`${url}/health`);
        if (res.ok) break;
      } catch { /* not up yet */ }
      if (Date.now() > deadline) throw new Error(`server not ready: ${url}`);
      await sleep(500);
    }
  }
}

const commonOpts = () => ({
  websocket: WebSocket,
  fetch: fetch,
  readableStream: ReadableStream,
  eventsource: EventSource,
});

// Run one connect -> subscribe -> publish -> receive -> [idle hold] -> publish
// cycle. The idle hold keeps the connection open with ONLY server pings flowing;
// if the proxy drops the connection, the SDK fires 'disconnected'/'connecting'
// and we FAIL - crucially we detect a drop even if the SDK silently reconnects
// (which a plain state check at the end would miss).
async function runCase({ label, endpoints, emulationEndpoint, holdMs = LONGEVITY_WAIT_MS }) {
  const channel = `test:room-${label}-${uniq()}`;
  const options = {
    ...commonOpts(),
    emulationEndpoint,
    minReconnectDelay: 200,
    maxReconnectDelay: 2000,
  };
  // Print the EXACT centrifuge-js config this case uses. A single-element
  // transports array cannot fall back to another transport - the negotiated
  // transport is asserted below to equal the one we asked for.
  const expectedTransport = endpoints[0].transport;
  console.log(`    cfg ${label}: new Centrifuge(${JSON.stringify(endpoints)}, {emulationEndpoint: ${JSON.stringify(emulationEndpoint)}, ...node-shims})`);
  const centrifuge = new Centrifuge(endpoints, options);

  const connected = deferred();
  let everConnected = false;
  let negotiatedTransport = null;
  const drops = []; // any drop/reconnect AFTER the first successful connect
  centrifuge.on('connected', (ctx) => { everConnected = true; negotiatedTransport = ctx.transport; connected.resolve(); });
  centrifuge.on('disconnected', (ctx) => {
    if (everConnected) drops.push(`disconnected: ${ctx.code} ${ctx.reason}`);
  });
  centrifuge.on('connecting', (ctx) => {
    if (everConnected) drops.push(`reconnecting: ${ctx.code} ${ctx.reason}`);
  });
  centrifuge.on('error', (ctx) => {
    centrifuge._lastError = ctx?.error?.message || JSON.stringify(ctx?.error || ctx);
  });

  const sub = centrifuge.newSubscription(channel);
  const subscribed = deferred();
  const inbox = new Map(); // token -> deferred
  const expect = (token) => {
    const d = deferred();
    inbox.set(token, d);
    return d.promise;
  };
  sub.on('subscribed', () => subscribed.resolve());
  sub.on('error', (ctx) => { centrifuge._lastSubError = JSON.stringify(ctx); });
  sub.on('publication', (ctx) => {
    const t = ctx.data && ctx.data.token;
    if (t && inbox.has(t)) inbox.get(t).resolve();
  });

  try {
    sub.subscribe();
    centrifuge.connect();

    await withTimeout(connected.promise, CONNECT_TIMEOUT_MS,
      `connect (${centrifuge._lastError || 'no error detail'})`);

    // Guarantee we actually used the transport under test - never a silent
    // fallback to WebSocket. For emulation transports this means /emulation is
    // the only client->server path.
    if (negotiatedTransport !== expectedTransport) {
      throw new Error(`transport mismatch: negotiated '${negotiatedTransport}', expected '${expectedTransport}'`);
    }

    await withTimeout(subscribed.promise, CONNECT_TIMEOUT_MS,
      `subscribe (${centrifuge._lastSubError || 'no error detail'})`);

    // 1) Delivery + no-buffering: publish and require prompt receipt.
    const t1 = `t1-${uniq()}`;
    const got1 = expect(t1);
    await apiPublish(channel, { token: t1 });
    await withTimeout(got1, RECEIVE_TIMEOUT_MS, 'first publication (buffering?)');

    // 2) Idle survival: hold with only pings flowing, then require delivery
    //    again. Any drop during the hold is a proxy timeout misconfiguration.
    await sleep(holdMs);
    if (drops.length > 0) {
      throw new Error(`connection dropped during ${Math.round(holdMs / 1000)}s idle hold -> ${drops.join('; ')}`);
    }
    if (centrifuge.state !== 'connected') {
      throw new Error(`not connected after idle hold (state=${centrifuge.state})`);
    }
    const t2 = `t2-${uniq()}`;
    const got2 = expect(t2);
    await apiPublish(channel, { token: t2 });
    await withTimeout(got2, RECEIVE_TIMEOUT_MS, 'second publication after idle');

    return { ok: true, transport: negotiatedTransport };
  } finally {
    try { centrifuge.disconnect(); } catch { /* ignore */ }
  }
}

function endpointsFor(transport, scheme) {
  const host = PROXY_HOST;
  const port = scheme === 'https' ? HTTPS_PORT : HTTP_PORT;
  const httpBase = `${scheme}://${host}:${port}`;
  const wsScheme = scheme === 'https' ? 'wss' : 'ws';
  const wsBase = `${wsScheme}://${host}:${port}`;
  const endpoint = transport === 'websocket'
    ? `${wsBase}/connection/websocket`
    : `${httpBase}/connection/${transport}`;
  return {
    endpoints: [{ transport, endpoint }],
    emulationEndpoint: `${httpBase}/emulation`,
  };
}

async function runMatrix() {
  const schemes = ['http', 'https'];
  const transports = ['websocket', 'http_stream', 'sse'];
  const results = [];
  for (const scheme of schemes) {
    for (const transport of transports) {
      const label = `${transport}/${scheme}`;
      const started = Date.now();
      try {
        const r = await runCase({ label, ...endpointsFor(transport, scheme) });
        results.push({ label, ok: true, ms: Date.now() - started });
        console.log(`  PASS  ${label}  [transport=${r.transport}]  (${Date.now() - started}ms)`);
      } catch (err) {
        results.push({ label, ok: false, ms: Date.now() - started, err: err.message });
        console.log(`  FAIL  ${label}  -> ${err.message}`);
      }
    }
  }
  return results;
}

// Deterministic cross-node emulation: stream on node-1, /emulation POST on
// node-2. Requires Centrifugo to route the surveyed command to the owning node.
async function runCrossNode() {
  const results = [];
  for (const transport of ['http_stream', 'sse']) {
    const label = `crossnode:${transport} (stream=node1, emulation=node2)`;
    const started = Date.now();
    try {
      const r = await runCase({
        label,
        endpoints: [{ transport, endpoint: `${NODE1_URL}/connection/${transport}` }],
        emulationEndpoint: `${NODE2_URL}/emulation`,
      });
      results.push({ label, ok: true, ms: Date.now() - started });
      console.log(`  PASS  ${label}  [transport=${r.transport}]  (${Date.now() - started}ms)`);
    } catch (err) {
      results.push({ label, ok: false, ms: Date.now() - started, err: err.message });
      console.log(`  FAIL  ${label}  -> ${err.message}`);
    }
  }
  return results;
}

// Idle-timeout probe: open all three transports through the proxy at once and
// hold them idle (only pings) for IDLE_HOLD_MS, asserting none drops. Run
// concurrently so one long window covers all transports.
async function runIdle() {
  const transports = ['websocket', 'http_stream', 'sse'];
  console.log(`  holding ${transports.length} idle connections for ${Math.round(IDLE_HOLD_MS / 1000)}s (only pings flow)...`);
  const jobs = transports.map(async (transport) => {
    const label = `idle:${transport}`;
    const started = Date.now();
    try {
      const r = await runCase({ label, ...endpointsFor(transport, 'http'), holdMs: IDLE_HOLD_MS });
      console.log(`  PASS  ${label}  [transport=${r.transport}]  (${Date.now() - started}ms)`);
      return { label, ok: true, ms: Date.now() - started };
    } catch (err) {
      console.log(`  FAIL  ${label}  -> ${err.message}`);
      return { label, ok: false, ms: Date.now() - started, err: err.message };
    }
  });
  return Promise.all(jobs);
}

// Cross-node PUB/SUB through the balancer: open many connections (mixed across
// all three transports), verify via /api/info that the balancer spread them over
// BOTH nodes, then publish several messages and require EVERY connection to
// receive EVERY message. This proves the broker fans a publication out across
// nodes and nothing is lost - regardless of which node a subscriber landed on.
async function runFanout() {
  const N = parseInt(process.env.FANOUT_CONNS || '9', 10);
  const M = parseInt(process.env.FANOUT_MSGS || '5', 10);
  const channel = `test:fanout-${uniq()}`;
  const transports = ['websocket', 'http_stream', 'sse'];
  const tokens = Array.from({ length: M }, (_, i) => `fan-${uniq()}-${i}`);
  const wanted = new Set(tokens);

  console.log(`  opening ${N} connections (mixed transports) through the proxy, all on ${channel}`);
  const conns = [];
  for (let i = 0; i < N; i++) {
    const transport = transports[i % transports.length];
    const { endpoints, emulationEndpoint } = endpointsFor(transport, 'http');
    const c = new Centrifuge(endpoints, { ...commonOpts(), emulationEndpoint, minReconnectDelay: 200, maxReconnectDelay: 2000 });
    const connected = deferred();
    const subscribed = deferred();
    const allMsgs = deferred();
    const got = new Set();
    let node = null;
    c.on('connected', (ctx) => { node = ctx.transport; connected.resolve(); });
    const sub = c.newSubscription(channel);
    sub.on('subscribed', () => subscribed.resolve());
    sub.on('publication', (ctx) => {
      const t = ctx.data && ctx.data.token;
      if (t && wanted.has(t)) { got.add(t); if (got.size === M) allMsgs.resolve(); }
    });
    conns.push({ i, transport, c, sub, connected, subscribed, allMsgs, got });
  }

  try {
    conns.forEach((x) => { x.sub.subscribe(); x.c.connect(); });
    await withTimeout(Promise.all(conns.map((x) => x.connected.promise)), CONNECT_TIMEOUT_MS, 'fanout connect');
    await withTimeout(Promise.all(conns.map((x) => x.subscribed.promise)), CONNECT_TIMEOUT_MS, 'fanout subscribe');

    // Distribution: confirm the balancer placed connections on BOTH nodes.
    // Per-node num_clients in /api/info refreshes on an interval (node pings),
    // so poll until it reflects our freshly-connected clients.
    let dist = '';
    let nodesWithClients = 0;
    const distDeadline = Date.now() + 15000;
    while (Date.now() < distDeadline) {
      const nodes = await apiInfo();
      const total = nodes.reduce((s, n) => s + (n.num_clients || 0), 0);
      nodesWithClients = nodes.filter((n) => (n.num_clients || 0) > 0).length;
      dist = nodes.map((n) => `${n.name}:${n.num_clients || 0}`).sort().join('  ');
      if (nodesWithClients >= 2 && total >= N) break;
      await sleep(500);
    }
    console.log(`  node distribution (via /api/info): ${dist}`);
    if (nodesWithClients < 2) {
      throw new Error(`connections did not spread across both nodes: ${dist}`);
    }

    // Publish M messages and require every connection to receive every one.
    for (const token of tokens) await apiPublish(channel, { token });
    await withTimeout(Promise.all(conns.map((x) => x.allMsgs.promise)), RECEIVE_TIMEOUT_MS,
      'fanout delivery (some connection missed a message)');

    console.log(`  all ${N} connections received all ${M} messages`);
    return [{ label: `fanout (${N} conns x ${M} msgs across 2 nodes)`, ok: true }];
  } catch (err) {
    const missing = conns.filter((x) => x.got.size < M).map((x) => `${x.transport}#${x.i}:${x.got.size}/${M}`);
    console.log(`  FAIL fanout -> ${err.message}${missing.length ? ` | missing: ${missing.join(', ')}` : ''}`);
    return [{ label: 'fanout', ok: false, err: err.message }];
  } finally {
    conns.forEach((x) => { try { x.c.disconnect(); } catch { /* ignore */ } });
  }
}

async function main() {
  console.log(`\n=== tester: mode=${MODE} proxy=${PROXY} ===`);
  await waitReady();
  let results;
  if (MODE === 'crossnode') results = await runCrossNode();
  else if (MODE === 'idle') results = await runIdle();
  else if (MODE === 'fanout') results = await runFanout();
  else results = await runMatrix();
  const failed = results.filter((r) => !r.ok);
  console.log(`\n=== ${MODE} ${PROXY}: ${results.length - failed.length}/${results.length} passed ===\n`);
  process.exit(failed.length === 0 ? 0 : 1);
}

main().catch((err) => { console.error('tester crashed:', err); process.exit(2); });
