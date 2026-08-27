#!/usr/bin/env node
// ponytail: minimal cross-platform postinstall — downloads cgo/meteolink binary for linux/darwin/win from GitHub releases
// doubles as bin shim: `meteolink top` spawns the downloaded Go binary
const fs=require('fs'), path=require('path'), https=require('https'), cp=require('child_process')
const REPO='Realms4239/cgo'
const BIN_DIR=path.join(__dirname,'..','bin')
const BIN_NAME=process.platform==='win32'?'meteolink.exe':'meteolink'
const BIN_PATH=path.join(BIN_DIR, BIN_NAME)

function assetNames(){
  const plat=process.platform==='win32'?'windows':process.platform==='darwin'?'darwin':'linux'
  const arch=process.arch==='arm64'?'arm64':process.arch==='x64'?'amd64':'amd64'
  const ext=plat==='windows'?'.exe':''
  return [
    `meteolink-${plat}-${arch}${ext}`,
    `cgo-${plat}-${arch}${ext}`,
    `meteolink-${plat}-${arch}`,
    `cgo-${plat}-${arch}`,
  ]
}

function download(url, dest){
  return new Promise((res,rej)=>{
    const f=fs.createWriteStream(dest)
    https.get(url, r=>{
      if(r.statusCode>=300 && r.statusCode<400 && r.headers.location){
        f.close(); fs.unlink(dest,()=>{}); download(r.headers.location, dest).then(res,rej); return
      }
      if(r.statusCode!==200){ f.close(); fs.unlink(dest,()=>{}); rej(new Error('HTTP '+r.statusCode+' '+url)); return }
      r.pipe(f); f.on('finish',()=>{ f.close(()=>{ try{fs.chmodSync(dest,0o755)}catch{}; res() }) }); f.on('error',rej)
    }).on('error', rej)
  })
}

async function postinstall(){
  const names=assetNames()
  if(fs.existsSync(BIN_PATH)){ console.log(`meteolink: binary exists ${BIN_PATH} — skip download`); return }
  fs.mkdirSync(BIN_DIR,{recursive:true})
  let lastErr=null
  for(const name of names){
    const url=`https://github.com/${REPO}/releases/latest/download/${name}`
    console.log(`meteolink: downloading ${url} -> ${BIN_PATH}`)
    try{ await download(url, BIN_PATH); console.log(`meteolink: installed ${BIN_PATH}`); return }catch(e){ lastErr=e; console.warn(`meteolink: miss ${name}: ${e.message}`) }
  }
  console.warn(`meteolink: download failed (${lastErr && lastErr.message})`)
  console.warn(`meteolink: fallback — go install github.com/${REPO}/cmd/meteolink@latest`)
  console.warn(`meteolink: or build locally — go build -o ${BIN_PATH} ./cmd/meteolink`)
}

function isBinInvocation(){
  const args=process.argv.slice(2)
  return args.length>0 && ['top','tui','--tui','-tui','--serve','-serve','serve','--web','--help','-h','--version'].some(a=>args.includes(a))
}

if(require.main===module){
  if(isBinInvocation()){
    if(!fs.existsSync(BIN_PATH)){
      console.error(`meteolink: binary not found at ${BIN_PATH}`)
      console.error(`meteolink: running postinstall to fetch it...`)
      postinstall().then(()=>{
        if(fs.existsSync(BIN_PATH)){
          const c=cp.spawn(BIN_PATH, process.argv.slice(2), {stdio:'inherit'})
          c.on('exit', code=>process.exit(code??0))
        } else process.exit(1)
      })
    } else {
      const c=cp.spawn(BIN_PATH, process.argv.slice(2), {stdio:'inherit'})
      c.on('exit', code=>process.exit(code??0)); c.on('error', e=>{ console.error('meteolink spawn:',e.message); process.exit(1) })
    }
  } else {
    postinstall().catch(e=>{ console.warn('meteolink postinstall:',e.message); process.exit(0) })
  }
} else {
  module.exports={postinstall}
}
