"""Run the handoff package checks and retain each attempt without overwriting it."""
import datetime
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys

root = Path('/Users/jacobdelgado/repos/spl-toolkit/.worktrees/remaining-milestones')
evidence = Path(__file__).resolve().parent
label = sys.argv[1]
record_path = evidence / (label + '.json')
assert not record_path.exists(), 'Choose a new attempt label'
env = os.environ.copy()
env.update(GOCACHE='/private/tmp/spl-toolkit-go-cache',
           GOMODCACHE='/private/tmp/spl-toolkit-handoff-gomodcache',
           GOPROXY='off', GOSUMDB='off', GOTOOLCHAIN='local')
command = ['/opt/homebrew/bin/go', 'test', '-mod=readonly', '-count=1',
           './pkg/rewrite', './pkg/analysis', './pkg/validation']

def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()

def sources():
    names = subprocess.check_output(
        ['git', 'ls-files', '--cached', '--others', '--exclude-standard', '-z',
         '*.go', 'go.mod', 'go.sum'], cwd=root).decode().split('\0')
    return {name: digest(root / name) for name in sorted(set(names)) if name}

record = {'command': command, 'cwd': str(root),
          'started_utc': datetime.datetime.now(datetime.timezone.utc).isoformat(),
          'head': subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=root, text=True).strip(),
          'go_version': subprocess.check_output([command[0], 'version'], env=env, text=True).strip(),
          'go_sha256': digest(Path(command[0]).resolve()),
          'environment': {key: env[key] for key in
                          ['GOCACHE', 'GOMODCACHE', 'GOPROXY', 'GOSUMDB', 'GOTOOLCHAIN']},
          'source_sha256': sources(),
          'scope': 'Single-query handoff checks, not complete Task5 or private acceptance.'}
record_path.write_text(json.dumps(record, indent=2) + '\n')
with (evidence / (label + '.stdout')).open('wb') as stdout, (evidence / (label + '.stderr')).open('wb') as stderr:
    result = subprocess.run(command, cwd=root, env=env, stdout=stdout, stderr=stderr)
record.update(exit_code=result.returncode,
              finished_utc=datetime.datetime.now(datetime.timezone.utc).isoformat(),
              source_unchanged=record['source_sha256'] == sources())
record['logs'] = {suffix: digest(evidence / (label + suffix)) for suffix in ['.stdout', '.stderr']}
record_path.write_text(json.dumps(record, indent=2) + '\n')
for suffix in ['.stdout', '.stderr']:
    print((evidence / (label + suffix)).read_bytes().decode('utf-8', errors='backslashreplace'), end='')
print('Observed exit:', result.returncode, 'Source unchanged:', record['source_unchanged'])
print('Receipt:', record_path)
sys.exit(result.returncode or (0 if record['source_unchanged'] else 1))
