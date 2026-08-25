export function PromptProgressLine({paused}:{paused:boolean}){
  return <div className="prompt-progress" style={{position:'absolute', bottom:0, left:0, height:1, background:'#f4b400', width:'100%', animation:'promptDrain 6000ms linear forwards', animationPlayState: paused?'paused':'running'}} />
}
