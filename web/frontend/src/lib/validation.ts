export function validate(data:any, rules:any){
  const errors:any={}
  for(const k in rules){
    const r=rules[k], v=data[k]
    if(r.required && (v==null || String(v).trim()==='')) errors[k]='requis'
    if(r.min!==undefined && v < r.min) errors[k]=`min ${r.min}`
    if(r.max!==undefined && v > r.max) errors[k]=`max ${r.max}`
  }
  return {valid: Object.keys(errors).length===0, errors}
}
