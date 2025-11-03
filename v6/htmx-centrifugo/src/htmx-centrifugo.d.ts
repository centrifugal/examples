/**
 * TypeScript type definitions for htmx-centrifugo
 * @version 0.1.0
 */

declare module 'htmx-centrifugo' {
  import { Centrifuge } from 'centrifuge';

  /**
   * Get the Centrifuge instance associated with an element
   */
  export function getCentrifugeInstance(element: Element): Centrifuge | null;

  /**
   * Cleanup Centrifugo connections and subscriptions for an element
   */
  export function cleanup(element: Element): void;
}

// Augment htmx types
declare global {
  interface HTMLElement {
    /**
     * WebSocket endpoint path or full URL
     * @example "/connection/websocket"
     * @example "ws://localhost:8000/connection/websocket"
     */
    'centrifugo-ws-endpoint'?: string;

    /**
     * HTTP-streaming endpoint path or full URL
     * @example "/connection/http_stream"
     */
    'centrifugo-http-stream-endpoint'?: string;

    /**
     * Server-Sent Events endpoint path or full URL
     * @example "/connection/sse"
     */
    'centrifugo-sse-endpoint'?: string;

    /**
     * Static JWT token for authentication
     */
    'centrifugo-token'?: string;

    /**
     * URL to fetch JWT token from
     * Server should return: { "token": "jwt-token-here" }
     */
    'centrifugo-token-url'?: string;

    /**
     * Enable debug logging
     */
    'centrifugo-debug'?: 'true' | 'false';

    /**
     * Enable connection init endpoint call (for HTTP/2 workaround)
     */
    'centrifugo-init'?: 'true' | 'false';

    /**
     * Subscribe to a Centrifugo channel
     * @example "news"
     * @example "chat:room123"
     */
    'centrifugo-subscribe'?: string;

    /**
     * URL to fetch subscription token from
     * Server should return: { "token": "jwt-token-here" }
     */
    'centrifugo-sub-token-url'?: string;

    /**
     * How to swap content when receiving updates
     */
    'centrifugo-swap'?:
      | 'innerHTML'
      | 'outerHTML'
      | 'beforebegin'
      | 'afterbegin'
      | 'beforeend'
      | 'afterend'
      | 'delete'
      | 'none';

    /**
     * CSS selector for target element to update
     * @example "#news-container"
     */
    'centrifugo-target'?: string;

    /**
     * Enable sending messages through this form
     */
    'centrifugo-send'?: '' | 'true';

    /**
     * Channel to publish to (when using centrifugo-send)
     */
    'centrifugo-channel'?: string;

    /**
     * Method to use for sending
     */
    'centrifugo-method'?: 'publish' | 'rpc';

    /**
     * RPC method name (when using centrifugo-method="rpc")
     */
    'centrifugo-rpc-method'?: string;
  }

  /**
   * Custom events fired by htmx-centrifugo
   */
  interface DocumentEventMap {
    'centrifugo:connecting': CustomEvent<CentrifugoConnectingEvent>;
    'centrifugo:connected': CustomEvent<CentrifugoConnectedEvent>;
    'centrifugo:disconnected': CustomEvent<CentrifugoDisconnectedEvent>;
    'centrifugo:error': CustomEvent<CentrifugoErrorEvent>;
    'centrifugo:subscribing': CustomEvent<CentrifugoSubscribingEvent>;
    'centrifugo:subscribed': CustomEvent<CentrifugoSubscribedEvent>;
    'centrifugo:unsubscribed': CustomEvent<CentrifugoUnsubscribedEvent>;
    'centrifugo:publication': CustomEvent<CentrifugoPublicationEvent>;
    'centrifugo:sub-error': CustomEvent<CentrifugoSubErrorEvent>;
    'centrifugo:sent': CustomEvent<CentrifugoSentEvent>;
    'centrifugo:send-error': CustomEvent<CentrifugoSendErrorEvent>;
    'centrifugo:rpc-result': CustomEvent<CentrifugoRpcResultEvent>;
  }

  interface CentrifugoConnectingEvent {
    code: number;
    reason: string;
  }

  interface CentrifugoConnectedEvent {
    client: string;
    transport: string;
    data?: any;
  }

  interface CentrifugoDisconnectedEvent {
    code: number;
    reason: string;
    reconnect: boolean;
  }

  interface CentrifugoErrorEvent {
    error: Error;
    type?: string;
  }

  interface CentrifugoSubscribingEvent {
    channel: string;
    ctx: {
      code: number;
      reason: string;
    };
  }

  interface CentrifugoSubscribedEvent {
    channel: string;
    ctx: {
      recoverable: boolean;
      positioned: boolean;
      data?: any;
      wasRecovering?: boolean;
      recovered?: boolean;
    };
  }

  interface CentrifugoUnsubscribedEvent {
    channel: string;
    ctx: {
      code: number;
      reason: string;
    };
  }

  interface CentrifugoPublicationEvent {
    channel: string;
    ctx: {
      data: any;
      offset?: number;
      tags?: Record<string, string>;
    };
  }

  interface CentrifugoSubErrorEvent {
    channel: string;
    ctx: {
      error: Error;
      type?: string;
    };
  }

  interface CentrifugoSentEvent {
    data: Record<string, any>;
  }

  interface CentrifugoSendErrorEvent {
    error: Error;
  }

  interface CentrifugoRpcResultEvent {
    method: string;
    result: {
      data?: any;
    };
  }
}

export {};
