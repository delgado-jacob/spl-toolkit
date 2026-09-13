const vscode = require('vscode');
const { LanguageClient } = require('vscode-languageclient/node');
let client;
async function activate() {
  const command = process.env.SPL_CONSUMER_CLI;
  if (!command) throw new Error('SPL_CONSUMER_CLI is required');
  client = new LanguageClient('spl-toolkit-acceptance', 'SPL Toolkit acceptance',
    { command, args: ['lsp', '--stdio'] },
    { documentSelector: [{ language: 'spl' }, { language: 'spl2' }] });
  await client.start();
  return { client };
}
async function deactivate() { if (client) await client.stop(); }
module.exports = { activate, deactivate };
