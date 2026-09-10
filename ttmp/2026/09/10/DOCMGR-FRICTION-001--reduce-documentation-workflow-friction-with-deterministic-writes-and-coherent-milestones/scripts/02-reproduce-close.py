"""Run with a pinned docmgr binary; all mutations occur in a temporary workspace."""
import hashlib,json,pathlib,re,subprocess,sys,tempfile
binary=str(pathlib.Path(sys.argv[1]).resolve())
with tempfile.TemporaryDirectory(prefix='docmgr-close-probe-') as root:
 def run(*args):
  p=subprocess.run([binary,*args],cwd=root,text=True,capture_output=True)
  return {'args':list(args),'exit':p.returncode,'stdout':p.stdout,'stderr':p.stderr}
 init=run('init','--seed-vocabulary');assert init['exit']==0,init
 results=[]
 for ticket in ['REPEAT-PROBE','FAIL-PROBE']:
  create=run('ticket','create','--ticket',ticket,'--title',ticket,'--topics','backend');assert create['exit']==0,create
  index=next(pathlib.Path(root).glob(f'ttmp/**/{ticket}--*/index.md'))
  log=index.parent/'changelog.md'
  if ticket=='FAIL-PROBE':
   log.unlink();log.mkdir() # Deterministic failure: changelog path cannot be opened as a file.
  first=run('ticket','close','--ticket',ticket)
  row={'ticket':ticket,'first':first,'status_after':re.search(r'^Status: (.*)$',index.read_text(),re.M)[1]}
  if ticket=='REPEAT-PROBE':
   before=index.read_bytes();second=run('ticket','close','--ticket',ticket)
   row.update(second=second,index_bytes_changed_again=before!=index.read_bytes(),
              closing_entries=log.read_text().count('Ticket closed'),
              trailing_newlines=len(log.read_bytes())-len(log.read_bytes().rstrip(b'\n')))
  results.append(row)
 print(json.dumps({'binary_sha256':hashlib.sha256(pathlib.Path(binary).read_bytes()).hexdigest(),'results':results},indent=2))
