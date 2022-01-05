module.exports = {
  branches: ['master'],
  plugins: [
    ['@semantic-release/commit-analyzer', { preset: 'conventionalcommits' }],
  ],
};