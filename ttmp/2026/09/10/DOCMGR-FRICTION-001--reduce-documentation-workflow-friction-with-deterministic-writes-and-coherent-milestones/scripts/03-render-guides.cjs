// node 03-render-guides.cjs PLAYWRIGHT_MODULE MERMAID_BUNDLE GUIDE...
// Source Markdown stays unchanged; derived print copies contain rendered PNGs.
const fs = require('node:fs');
const path = require('node:path');
const {chromium} = require(process.argv[2]);
(async () => {
 const browser = await chromium.launch({headless:true});
 try {
  const page = await browser.newPage({viewport:{width:1000,height:1400}});
  for (const file of process.argv.slice(4)) {
   const ticket = path.dirname(path.dirname(file));
   const out = path.join(ticket,'sources','print');fs.mkdirSync(out,{recursive:true});
   const input = fs.readFileSync(file,'utf8');
   const blocks = [...input.matchAll(/```mermaid\n([\s\S]*?)\n```/g)];
   const title=input.match(/^Title: (.+)$/m)[1];
   let printed=input.replace(/^---\n/,'---\ntitle: '+JSON.stringify(title)+'\n');const figures=[];
   for (let i=0;i<blocks.length;i++) {
    await page.setContent('<html><body style="margin:0;background:white"><div id="diagram"></div></body></html>');
    await page.addScriptTag({path:process.argv[3]});
    await page.evaluate(async ({text,id})=>{
     mermaid.initialize({startOnLoad:false,securityLevel:'strict',theme:'neutral'});
     const {svg}=await mermaid.render(id,text);
     document.querySelector('#diagram').innerHTML=svg;
    },{text:blocks[i][1],id:'figure'+i});
    const image=path.join(out,'figure-'+(i+1)+'.png');
    await page.locator('#diagram svg').screenshot({path:image});
    printed=printed.replace(blocks[i][0],`![](${image})`);figures.push(image);
   }
   fs.writeFileSync(path.join(out,'guide.md'),printed);
   const diary=fs.readFileSync(path.join(ticket,'reference','01-investigation-diary.md'),'utf8');
   fs.writeFileSync(path.join(out,'diary.md'),diary.replace(/^---\n/,'---\ntitle: "Investigation diary"\n'));
   fs.writeFileSync(path.join(out,'render.json'),JSON.stringify({source:file,figures,browser:await browser.version(),mermaid:'11.12.0',source_unchanged:true},null,2)+'\n');
   console.log(JSON.stringify({ticket,figures:figures.length,print:path.join(out,'guide.md')}));
  }
 } finally {await browser.close();}
})().catch(e=>{console.error(e);process.exitCode=1;});
