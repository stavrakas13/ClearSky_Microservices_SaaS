export function flash(msg){
  const d=document.createElement("div");
  d.textContent=msg;
  d.style.cssText="position:fixed;top:1rem;right:1rem;background:#006dd0;color:#fff;padding:0.5rem 1rem;border-radius:4px;z-index:1000";
  document.body.appendChild(d); setTimeout(()=>d.remove(),3e3);
}
