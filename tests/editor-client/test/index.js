const assert = require('assert/strict');
const fs = require('fs');
const vscode = require('vscode');

const range = r => [r.start.line, r.start.character, r.end.line, r.end.character];
const diagnostics = uri => vscode.languages.getDiagnostics(uri).map(d => ({
  code: typeof d.code === 'object' ? d.code.value : d.code, message: d.message,
  severity: d.severity, range: range(d.range)
}));
function changed(uri, predicate) {
  return new Promise((resolve, reject) => {
    const timeout = setTimeout(() => { subscription.dispose(); reject(new Error('diagnostic delivery timed out')); }, 20000);
    const subscription = vscode.languages.onDidChangeDiagnostics(event => {
      if (!event.uris.some(item => item.toString() === uri.toString())) return;
      const values = diagnostics(uri);
      if (predicate(values)) { clearTimeout(timeout); subscription.dispose(); resolve(values); }
    });
  });
}
async function replace(document, text) {
  const edit = new vscode.WorkspaceEdit();
  edit.replace(document.uri, new vscode.Range(document.positionAt(0), document.positionAt(document.getText().length)), text);
  assert.equal(await vscode.workspace.applyEdit(edit), true);
}
async function highlights(document, column) {
  const values = await vscode.commands.executeCommand('vscode.executeDocumentHighlights', document.uri, new vscode.Position(0, column));
  return (values || []).map(h => ({ kind: h.kind, range: range(h.range) }));
}
async function run() {
  const evidence = { status: 'failed', editor_version: vscode.version, client_version: require('vscode-languageclient/package.json').version, assertions: [] };
  try {
    const uri = vscode.Uri.file(process.env.SPL_CONSUMER_QUERY);
    const ready = changed(uri, d => d.some(item => item.code === 'SPL_SYNTAX_ERROR'));
    const document = await vscode.workspace.openTextDocument(uri);
    await vscode.window.showTextDocument(document);
    evidence.diagnostics = await ready;
    assert(evidence.diagnostics.some(d => d.severity === vscode.DiagnosticSeverity.Error && d.message && d.range[1] === "eval '😀a'=".length));
    evidence.assertions.push('delivered syntax error uses UTF16 EOF position');
    const firstVersion = document.version;
    const cleared = changed(uri, d => d.length === 0);
    await replace(document, 'eval a=1 | table a | eval a=2 | table a');
    await cleared;
    assert(document.version > firstVersion);
    evidence.versions = [firstVersion, document.version];
    evidence.assertions.push('full-text edit increments version and clears diagnostics');
    evidence.assignment_highlights = await highlights(document, 5);
    assert.deepEqual(evidence.assignment_highlights.map(h => h.range[1]), [5, 17]);
    assert(evidence.assignment_highlights.some(h => h.kind === vscode.DocumentHighlightKind.Write));
    assert(evidence.assignment_highlights.some(h => h.kind === vscode.DocumentHighlightKind.Read));
    evidence.assertions.push('registered provider separates independent assignments and retains read/write kinds');
    await replace(document, 'search root=1 | append [ search child=1 | eval root=child ] | where root=2');
    evidence.scope_highlights = await highlights(document, 7);
    assert.deepEqual(evidence.scope_highlights.map(h => h.range[1]), [7, 68]);
    evidence.assertions.push('same spelling in independent child scope is excluded');
    await replace(document, "eval '😀a'=host | table '😀a'");
    evidence.unicode_highlights = await highlights(document, 8);
    assert.equal(evidence.unicode_highlights[0].range[1], 5);
    assert.equal(evidence.unicode_highlights[0].range[3], 10);
    assert.equal(evidence.unicode_highlights[2].range[1], 24);
    evidence.assertions.push('non-BMP highlight ranges use UTF16');
    const broken = changed(uri, d => d.some(item => item.code === 'SPL_SYNTAX_ERROR'));
    await replace(document, 'eval a=');
    await broken;
    evidence.assertions.push('later invalid edit replaces diagnostics');
    const closed = changed(uri, d => d.length === 0);
    // Changing language emits a real didClose for the old document. Closing a
    // tab alone may leave VS Code's text model open in its document cache.
    await vscode.languages.setTextDocumentLanguage(document, 'plaintext');
    await closed;
    evidence.assertions.push('leaving SPL language closes client document and clears diagnostics');
    evidence.status = 'passed';
  } catch (error) { evidence.error = String(error.stack || error); throw error; }
  finally {
    const extension = vscode.extensions.all.find(item => item.packageJSON.name === 'spl-toolkit-consumer-acceptance');
    if (extension && extension.isActive && extension.exports.client) await extension.exports.client.stop();
    fs.writeFileSync(process.env.SPL_CONSUMER_RESULT, JSON.stringify(evidence, null, 2));
  }
}
module.exports = { run };
