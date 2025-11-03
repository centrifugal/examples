import resolve from '@rollup/plugin-node-resolve';
import terser from '@rollup/plugin-terser';

export default [
  // UMD build (for browser via script tag)
  {
    input: 'src/htmx-centrifugo.js',
    output: {
      file: 'dist/htmx-centrifugo.js',
      format: 'umd',
      name: 'htmxCentrifugo',
      globals: {
        centrifuge: 'Centrifuge'
      }
    },
    external: ['centrifuge'],
    plugins: [resolve()]
  },
  // Minified UMD build
  {
    input: 'src/htmx-centrifugo.js',
    output: {
      file: 'dist/htmx-centrifugo.min.js',
      format: 'umd',
      name: 'htmxCentrifugo',
      globals: {
        centrifuge: 'Centrifuge'
      }
    },
    external: ['centrifuge'],
    plugins: [resolve(), terser()]
  },
  // ESM build (for bundlers)
  {
    input: 'src/htmx-centrifugo.js',
    output: {
      file: 'dist/htmx-centrifugo.esm.js',
      format: 'es'
    },
    external: ['centrifuge'],
    plugins: [resolve()]
  }
];
