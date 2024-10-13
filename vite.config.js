export default {
  server: {
    proxy: {
      '/ws': {
        target: 'ws://localhost:8567',
        ws: true,
      },
    },
  },
};