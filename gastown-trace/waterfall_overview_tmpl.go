package main

const tmplWaterfallOverview = `
{{define "content"}}
<style>
/* ── Waterfall Overview — canvas swim-lane layout ────────────────────────── */
.ov-toolbar{display:flex;gap:8px;align-items:center;flex-wrap:wrap;padding:4px 0 8px}
.ov-toolbar select{background:#161b22;border:1px solid #30363d;color:#c9d1d9;padding:4px 8px;border-radius:4px;font-size:12px;font-family:monospace}
.ov-toolbar label{font-size:11px;color:#8b949e}
.ov-toolbar button{background:#21262d;border:1px solid #30363d;color:#c9d1d9;padding:4px 12px;border-radius:4px;cursor:pointer;font-size:12px}
.ov-cards{display:flex;gap:8px;flex-wrap:wrap;margin-bottom:8px}
.ov-card{background:#161b22;border:1px solid #30363d;border-radius:6px;padding:10px 16px;min-width:90px}
.ov-card-val{font-size:20px;font-weight:700;color:#c9d1d9;font-family:monospace}
.ov-card-lbl{font-size:10px;color:#8b949e;text-transform:uppercase;letter-spacing:.5px}
.ov-instance{font-size:11px;color:#58a6ff;margin-bottom:6px}
/* Mini overview */
#ov-mini-wrap{border:1px solid #30363d;border-radius:4px;background:#0d1117;margin-bottom:6px;cursor:crosshair;overflow:hidden;user-select:none;position:relative}
/* Main flex area */
#ov-main{display:flex;gap:0;border:1px solid #30363d;border-radius:4px;overflow:hidden;height:60vh;min-height:300px}
/* Labels column */
#ov-lab-wrap{width:160px;flex-shrink:0;overflow:hidden;position:relative;border-right:1px solid #30363d}
#ov-lab-canvas{display:block}
/* Swim lanes canvas */
#ov-canvas-wrap{flex:1;overflow:auto;position:relative;background:#0d1117}
#ov-canvas{display:block}
/* Detail panel */
#ov-detail{display:none;flex-direction:column;width:360px;min-width:280px;flex-shrink:0;border-left:1px solid #30363d;overflow:hidden}
#ov-detail-hdr{display:flex;justify-content:space-between;align-items:center;padding:8px 12px;background:#161b22;border-bottom:1px solid #30363d;font-size:12px;flex-shrink:0}
#ov-detail-hdr button{background:#21262d;border:1px solid #30363d;color:#c9d1d9;padding:2px 8px;border-radius:4px;cursor:pointer;font-size:11px}
#ov-detail-body{overflow-y:auto;flex:1;min-height:0;padding:8px;background:#0d1117}
.ov-dsect{font-size:9px;font-weight:700;letter-spacing:.8px;text-transform:uppercase;color:#58a6ff;padding:6px 0 3px;border-bottom:1px solid #1c2128;margin-bottom:4px}
.ov-drow{display:flex;gap:8px;padding:2px 0;border-bottom:1px solid #0d1117}
.ov-dkey{color:#8b949e;min-width:76px;flex-shrink:0;font-size:10px}
.ov-dval{color:#c9d1d9;font-family:monospace;font-size:10px;word-break:break-all}
.ov-evlist{margin-top:4px}
.ov-evitem{padding:3px 6px;border-left:2px solid #30363d;margin-bottom:2px;font-size:10px;cursor:default;display:flex;gap:6px;align-items:baseline}
.ov-evtime{color:#8b949e;flex-shrink:0}
.ov-evbody{font-family:monospace}
/* Tooltip */
#ov-tip{position:fixed;background:#1c2128;border:1px solid #30363d;padding:8px 10px;border-radius:4px;font-size:11px;color:#c9d1d9;pointer-events:none;display:none;z-index:200;max-width:280px;line-height:1.6;box-shadow:0 4px 12px rgba(0,0,0,.5)}
</style>

{{if .Instance}}<div class="ov-instance">instance: <b>{{.Instance}}</b>{{if .TownRoot}} &nbsp;·&nbsp; {{.TownRoot}}{{end}}</div>{{end}}

<div class="ov-toolbar">
  <label>Rig</label>
  <select id="f-rig" onchange="applyFilters()"><option value="">All rigs</option></select>
  <label>Role</label>
  <select id="f-role" onchange="applyFilters()">
    <option value="">All roles</option>
    <option>mayor</option><option>deacon</option><option>witness</option>
    <option>refinery</option><option>polecat</option><option>dog</option>
    <option>boot</option><option>crew</option>
  </select>
  <button onclick="resetFilters()">Reset</button>
  <span style="flex:1"></span>
  <span style="font-size:11px;color:#8b949e">window: {{.Window}}</span>
</div>

<div class="ov-cards">
  <div class="ov-card"><div class="ov-card-val" id="s-runs">—</div><div class="ov-card-lbl">Runs</div></div>
  <div class="ov-card"><div class="ov-card-val" id="s-rigs">—</div><div class="ov-card-lbl">Rigs</div></div>
  <div class="ov-card"><div class="ov-card-val" id="s-events">—</div><div class="ov-card-lbl">Events</div></div>
  <div class="ov-card"><div class="ov-card-val" id="s-cost">—</div><div class="ov-card-lbl">Cost</div></div>
  <div class="ov-card"><div class="ov-card-val" id="s-dur">—</div><div class="ov-card-lbl">Duration</div></div>
</div>

<div id="ov-mini-wrap"><canvas id="ov-mini"></canvas></div>
<div id="ov-tip"></div>

<div id="ov-main">
  <div id="ov-lab-wrap"><canvas id="ov-lab-canvas"></canvas></div>
  <div id="ov-canvas-wrap"><canvas id="ov-canvas"></canvas></div>
  <div id="ov-detail">
    <div id="ov-detail-hdr">
      <span id="ov-detail-title">Run detail</span>
      <button onclick="closeDetail()">✕</button>
    </div>
    <div id="ov-detail-body"></div>
  </div>
</div>

<script>
(function(){
var DATA = {{.JSONData}};

// ── Constants ─────────────────────────────────────────────────────────────────
var ROLE_CLR = {
  mayor:'#f59e0b', deacon:'#8b5cf6', witness:'#3b82f6',
  refinery:'#10b981', polecat:'#ef4444', dog:'#f97316',
  boot:'#6b7280', crew:'#06b6d4'
};
var EV_CLR = {
  'agent.instantiate':'#06b6d4','session.start':'#06b6d4','session.stop':'#56d364',
  'prime':'#3b82f6','prompt.send':'#06b6d4','agent.event':'#8b5cf6',
  'bd.call':'#3b82f6','claude_code.api_request':'#f59e0b',
  'claude_code.tool_result':'#10b981',
  'sling':'#eab308','mail':'#eab308','nudge':'#8b5cf6',
  'polecat.spawn':'#ef4444','polecat.remove':'#ef4444',
  'done':'#56d364','formula.instantiate':'#3b82f6','convoy.create':'#3b82f6'
};
var COMM_CLR = { sling:'#eab308', mail:'#06b6d4', nudge:'#8b5cf6', spawn:'#ef4444', done:'#56d364' };
var LANE_H = 30;
var RIG_H  = 22;
var LABEL_W = 160;
var TICK_W  = 2;

function rc(r){ return ROLE_CLR[r]||'#9ca3af'; }
function ec(b){ return EV_CLR[b]||'#30363d'; }

// ── State ─────────────────────────────────────────────────────────────────────
var filteredRigs=[], lanes=[], runY={}, totalH=0;
var fullStart=0, fullEnd=0, vStart=0, vEnd=0;
var miniC, miniCtx, mainC, mainCtx, labC, labCtx;
var detPanel, detBody, detTitle, tip;
var miniDrag=false, miniMode=null, miniX0=0, miniVS0=0, miniVE0=0;
var selectedRun=null;

// ── Init ──────────────────────────────────────────────────────────────────────
function init(){
  miniC    = document.getElementById('ov-mini');
  miniCtx  = miniC.getContext('2d');
  mainC    = document.getElementById('ov-canvas');
  mainCtx  = mainC.getContext('2d');
  labC     = document.getElementById('ov-lab-canvas');
  labCtx   = labC.getContext('2d');
  detPanel = document.getElementById('ov-detail');
  detBody  = document.getElementById('ov-detail-body');
  detTitle = document.getElementById('ov-detail-title');
  tip      = document.getElementById('ov-tip');

  var mWrap = document.getElementById('ov-mini-wrap');
  miniC.width = mWrap.clientWidth || 900;
  miniC.height = 56;
  mWrap.style.height = '56px';

  // Populate rig filter
  var selRig = document.getElementById('f-rig');
  for(var i=0;i<DATA.rigs.length;i++){
    var o=document.createElement('option');
    o.value=o.textContent=DATA.rigs[i].name;
    selRig.appendChild(o);
  }
  // Restore URL params
  var p=new URLSearchParams(window.location.search);
  if(p.get('rig'))  document.getElementById('f-rig').value  = p.get('rig');
  if(p.get('role')) document.getElementById('f-role').value = p.get('role');

  applyFiltersLocal();
  buildLanes();
  computeRange();
  vStart=fullStart; vEnd=fullEnd;
  updateSummary();
  resize();

  miniC.addEventListener('mousedown', onMiniDown);
  miniC.addEventListener('mousemove', onMiniHover);
  miniC.addEventListener('mouseleave', function(){ miniC.style.cursor='crosshair'; });
  window.addEventListener('mousemove', onWinMove);
  window.addEventListener('mouseup',   onWinUp);
  window.addEventListener('resize',    resize);
  window.addEventListener('keydown',   function(e){ if(e.key==='Escape') closeDetail(); });

  mainC.addEventListener('click',      onMainClick);
  mainC.addEventListener('mousemove',  onMainHover);
  mainC.addEventListener('mouseleave', function(){ tip.style.display='none'; });

  // Sync label scroll with main scroll
  var cWrap = document.getElementById('ov-canvas-wrap');
  cWrap.addEventListener('scroll', function(){
    document.getElementById('ov-lab-wrap').scrollTop = cWrap.scrollTop;
    drawMain(); drawLabels();
  });
}

// ── Filters ───────────────────────────────────────────────────────────────────
function applyFiltersLocal(){
  var rigF  = document.getElementById('f-rig').value;
  var roleF = document.getElementById('f-role').value;
  filteredRigs = DATA.rigs.map(function(rig){
    if(rigF && rig.name !== rigF) return null;
    var runs = rig.runs.filter(function(r){ return !roleF || r.role === roleF; });
    return runs.length ? {name:rig.name, runs:runs} : null;
  }).filter(Boolean);
}

function buildLanes(){
  lanes=[]; runY={}; totalH=0;
  for(var i=0;i<filteredRigs.length;i++){
    var rig=filteredRigs[i];
    lanes.push({type:'rig', rig:rig, y:totalH});
    totalH+=RIG_H;
    for(var j=0;j<rig.runs.length;j++){
      var run=rig.runs[j];
      lanes.push({type:'run', rig:rig, run:run, y:totalH});
      runY[run.run_id]={y:totalH+LANE_H/2, rig:rig.name};
      totalH+=LANE_H;
    }
  }
  totalH+=8;
}

function computeRange(){
  var mn=Infinity, mx=-Infinity;
  for(var i=0;i<lanes.length;i++){
    if(lanes[i].type!=='run') continue;
    var r=lanes[i].run;
    var st=new Date(r.started_at).getTime(); if(st<mn) mn=st;
    var et=r.ended_at?new Date(r.ended_at).getTime():Date.now(); if(et>mx) mx=et;
  }
  if(!isFinite(mn)){ mn=Date.now()-3600000; mx=Date.now(); }
  var pad=(mx-mn)*0.04||5000;
  fullStart=mn-pad; fullEnd=mx+pad;
}

// ── Resize ────────────────────────────────────────────────────────────────────
function resize(){
  var cWrap = document.getElementById('ov-canvas-wrap');
  var W = cWrap.clientWidth || 800;
  var H = Math.max(totalH, 200);
  mainC.width=W; mainC.height=H;
  labC.width=LABEL_W; labC.height=H;
  document.getElementById('ov-lab-wrap').style.overflowY='hidden';

  var mWrap = document.getElementById('ov-mini-wrap');
  miniC.width = mWrap.clientWidth || 900;

  drawMini(); drawMain(); drawLabels();
}

// ── Time helpers ──────────────────────────────────────────────────────────────
function tX(t, W){ return ((t-vStart)/(vEnd-vStart))*W; }
function miniX(t, W){ return ((t-fullStart)/(fullEnd-fullStart))*W; }

// ── Mini canvas ───────────────────────────────────────────────────────────────
function drawMini(){
  var W=miniC.width, H=miniC.height;
  miniCtx.fillStyle='#0d1117';
  miniCtx.fillRect(0,0,W,H);

  // All runs from full DATA (unfiltered for full picture)
  var allRuns=[];
  for(var i=0;i<DATA.rigs.length;i++)
    for(var j=0;j<DATA.rigs[i].runs.length;j++)
      allRuns.push(DATA.rigs[i].runs[j]);

  var barZone=H-14;
  var lH=allRuns.length?Math.max(1.5,Math.min(8,barZone/allRuns.length)):4;
  for(var k=0;k<allRuns.length;k++){
    var run=allRuns[k];
    var x1=miniX(new Date(run.started_at).getTime(),W);
    var x2=Math.max(x1+1,miniX(run.ended_at?new Date(run.ended_at).getTime():fullEnd,W));
    var y2=2+k*lH; if(y2>barZone) break;
    miniCtx.fillStyle=rc(run.role)+'99';
    miniCtx.fillRect(x1,y2,x2-x1,Math.max(1,lH-0.5));
  }

  // Dim outside selection + border + handles
  var sx1=miniX(vStart,W), sx2=miniX(vEnd,W);
  miniCtx.fillStyle='rgba(0,0,0,0.55)';
  if(sx1>0) miniCtx.fillRect(0,0,Math.max(0,sx1),H);
  if(sx2<W) miniCtx.fillRect(sx2,0,W-sx2,H);
  miniCtx.strokeStyle='#58a6ff'; miniCtx.lineWidth=1;
  miniCtx.strokeRect(sx1+.5,.5,Math.max(4,sx2-sx1)-1,H-1);
  miniCtx.fillStyle='#58a6ff';
  miniCtx.fillRect(sx1-1,0,3,H);
  miniCtx.fillRect(sx2-2,0,3,H);

  // Ruler
  var iv=niceInterval(fullEnd-fullStart,W), first=Math.ceil(fullStart/iv)*iv;
  miniCtx.font='8px monospace'; miniCtx.textAlign='center';
  for(var t=first;t<=fullEnd;t+=iv){
    var x=miniX(t,W);
    miniCtx.fillStyle='#30363d'; miniCtx.fillRect(x,H-14,1,4);
    miniCtx.fillStyle='#8b949e'; miniCtx.fillText(rulerLabel(t),x,H-1);
  }
}

// ── Main swim-lane canvas ─────────────────────────────────────────────────────
function drawMain(){
  var W=mainC.width, H=mainC.height;
  mainCtx.fillStyle='#0d1117';
  mainCtx.fillRect(0,0,W,H);

  // Grid lines + ruler
  var iv=niceInterval(vEnd-vStart,W), first=Math.ceil(vStart/iv)*iv;
  mainCtx.font='9px monospace'; mainCtx.textAlign='center';
  for(var t=first;t<=vEnd;t+=iv){
    var rx=tX(t,W);
    mainCtx.fillStyle='#161b22';
    mainCtx.fillRect(rx,0,1,H);
    mainCtx.fillStyle='#8b949e';
    mainCtx.fillText(rulerLabel(t),rx,10);
  }

  // Lanes
  for(var i=0;i<lanes.length;i++){
    var lane=lanes[i];
    if(lane.type==='rig'){
      mainCtx.fillStyle='#161b22';
      mainCtx.fillRect(0,lane.y,W,RIG_H);
      mainCtx.fillStyle='#30363d';
      mainCtx.fillRect(0,lane.y+RIG_H-1,W,1);
    } else {
      var run=lane.run, y=lane.y;
      var st=new Date(run.started_at).getTime();
      var et=run.ended_at?new Date(run.ended_at).getTime():fullEnd;
      var x1=tX(st,W), x2=Math.max(x1+4,tX(et,W));
      var isSelected=(run===selectedRun);

      // Lane background
      mainCtx.fillStyle=isSelected?'#1f6feb22':'#161b2211';
      mainCtx.fillRect(0,y,W,LANE_H);

      // Run bar
      var barClr=rc(run.role);
      mainCtx.fillStyle=barClr+(isSelected?'88':'44');
      mainCtx.fillRect(x1,y+7,x2-x1,LANE_H-14);
      mainCtx.strokeStyle=barClr;
      mainCtx.lineWidth=isSelected?2:1;
      mainCtx.strokeRect(x1+.5,y+7.5,Math.max(3,x2-x1-1),LANE_H-15);

      // Event ticks
      var evs=run.events||[];
      for(var k=0;k<evs.length;k++){
        var ev=evs[k];
        var ex=tX(new Date(ev.timestamp).getTime(),W);
        if(ex<0||ex>W) continue;
        mainCtx.fillStyle=ec(ev.body);
        mainCtx.fillRect(ex,y+9,TICK_W,LANE_H-18);
      }

      // Row separator
      mainCtx.fillStyle='#21262d';
      mainCtx.fillRect(0,y+LANE_H-1,W,1);
    }
  }

  // Comm arrows
  drawComms(W);
}

function drawComms(W){
  if(!DATA.communications) return;
  for(var i=0;i<DATA.communications.length;i++){
    var c=DATA.communications[i];
    if(c.type!=='sling'&&c.type!=='mail'&&c.type!=='nudge') continue;
    var fy=runY[c.from], ty=runY[c.to];
    if(!fy||!ty) continue;
    var ct=tX(new Date(c.time).getTime(),W);
    if(ct<0||ct>W) continue;
    var clr=(COMM_CLR[c.type]||'#8b949e')+'cc';
    mainCtx.strokeStyle=clr; mainCtx.lineWidth=1;
    mainCtx.setLineDash([3,3]);
    mainCtx.beginPath();
    mainCtx.moveTo(ct,fy.y);
    mainCtx.lineTo(ct,ty.y);
    mainCtx.stroke();
    mainCtx.setLineDash([]);
    // Arrowhead
    var dy=ty.y>fy.y?4:-4;
    mainCtx.fillStyle=COMM_CLR[c.type]||'#8b949e';
    mainCtx.beginPath();
    mainCtx.moveTo(ct,ty.y);
    mainCtx.lineTo(ct-3,ty.y-dy);
    mainCtx.lineTo(ct+3,ty.y-dy);
    mainCtx.closePath();
    mainCtx.fill();
  }
}

// ── Labels canvas ─────────────────────────────────────────────────────────────
function drawLabels(){
  var W=LABEL_W, H=labC.height;
  labCtx.fillStyle='#0d1117';
  labCtx.fillRect(0,0,W,H);
  labCtx.fillStyle='#30363d';
  labCtx.fillRect(W-1,0,1,H);

  for(var i=0;i<lanes.length;i++){
    var lane=lanes[i];
    if(lane.type==='rig'){
      labCtx.fillStyle='#161b22';
      labCtx.fillRect(0,lane.y,W,RIG_H);
      labCtx.fillStyle='#8b949e';
      labCtx.font='bold 10px monospace';
      labCtx.textAlign='left';
      labCtx.fillText(lane.rig.name.toUpperCase(), 8, lane.y+RIG_H/2+4);
      labCtx.fillStyle='#30363d';
      labCtx.fillRect(0,lane.y+RIG_H-1,W,1);
    } else {
      var run=lane.run, y=lane.y;
      var isSelected=(run===selectedRun);
      labCtx.fillStyle=isSelected?'#1f6feb22':'transparent';
      labCtx.fillRect(0,y,W,LANE_H);
      // Role dot
      labCtx.fillStyle=rc(run.role);
      labCtx.beginPath();
      labCtx.arc(10,y+LANE_H/2,4,0,2*Math.PI);
      labCtx.fill();
      // Agent name
      labCtx.fillStyle='#c9d1d9';
      labCtx.font='11px monospace';
      labCtx.textAlign='left';
      labCtx.fillText((run.agent_name||run.role||'').substring(0,14), 20, y+LANE_H/2+4);
      // Cost (below agent name)
      if(run.cost>0){
        labCtx.fillStyle='#8b949e';
        labCtx.font='9px monospace';
        labCtx.fillText('$'+run.cost.toFixed(3), 20, y+LANE_H/2+14);
      }
      labCtx.fillStyle='#21262d';
      labCtx.fillRect(0,y+LANE_H-1,W,1);
    }
  }
}

// ── Interactions ──────────────────────────────────────────────────────────────
function laneAt(y){
  for(var i=0;i<lanes.length;i++){
    var l=lanes[i], h=(l.type==='rig')?RIG_H:LANE_H;
    if(y>=l.y&&y<l.y+h) return l;
  }
  return null;
}

function onMainClick(e){
  var r=mainC.getBoundingClientRect();
  var lane=laneAt(e.clientY-r.top);
  if(!lane||lane.type!=='run') return;
  selectedRun=lane.run;
  showDetail(lane.run);
  drawMain(); drawLabels();
}

function onMainHover(e){
  var r=mainC.getBoundingClientRect();
  var lane=laneAt(e.clientY-r.top);
  if(!lane||lane.type!=='run'){ tip.style.display='none'; return; }
  var run=lane.run;
  tip.innerHTML='<b style="color:'+rc(run.role)+'">'+esc(run.role)+'</b>'
    +' <span style="color:#8b949e">'+esc(run.agent_name||'')+'</span><br>'
    +'Rig: '+esc(run.rig||'—')+'<br>'
    +'Events: '+((run.events||[]).length)+'<br>'
    +'Duration: '+(run.duration_ms?fmtMs(run.duration_ms):run.running?'running':'?')+'<br>'
    +(run.cost?'Cost: $'+run.cost.toFixed(4):'');
  tip.style.left=(e.clientX+12)+'px';
  tip.style.top=(e.clientY-10)+'px';
  tip.style.display='block';
}

// ── Mini interactions ─────────────────────────────────────────────────────────
function onMiniHover(e){
  var W=miniC.width, x=e.clientX-miniC.getBoundingClientRect().left;
  var sx1=miniX(vStart,W), sx2=miniX(vEnd,W);
  miniC.style.cursor=
    (Math.abs(x-sx1)<8||Math.abs(x-sx2)<8)?'ew-resize':
    (x>sx1&&x<sx2)?'grab':'crosshair';
}

function onMiniDown(e){
  var W=miniC.width, x=e.clientX-miniC.getBoundingClientRect().left;
  var sx1=miniX(vStart,W), sx2=miniX(vEnd,W);
  if(Math.abs(x-sx1)<8)      miniMode='left';
  else if(Math.abs(x-sx2)<8) miniMode='right';
  else if(x>sx1&&x<sx2)      miniMode='pan';
  else {
    var t=fullStart+(x/W)*(fullEnd-fullStart), half=(vEnd-vStart)/2;
    vStart=t-half; vEnd=t+half;
    drawMini(); drawMain(); return;
  }
  miniDrag=true; miniX0=x; miniVS0=vStart; miniVE0=vEnd;
  miniC.style.cursor='grabbing';
  e.preventDefault();
}

function onWinMove(e){
  if(!miniDrag) return;
  var W=miniC.width, x=e.clientX-miniC.getBoundingClientRect().left;
  var dt=(x-miniX0)/W*(fullEnd-fullStart);
  if(miniMode==='pan'){ vStart=miniVS0+dt; vEnd=miniVE0+dt; }
  else if(miniMode==='left')  vStart=Math.min(miniVS0+dt, miniVE0-1000);
  else                        vEnd  =Math.max(miniVE0+dt, miniVS0+1000);
  drawMini(); drawMain();
}

function onWinUp(){ if(miniDrag){ miniDrag=false; miniC.style.cursor='crosshair'; } }

// ── Detail panel ──────────────────────────────────────────────────────────────
function showDetail(run){
  detTitle.textContent=(run.agent_name||run.role||'Run')+' · '+(run.rig||'town');
  var h='';
  h+='<div class="ov-dsect">Run</div>';
  h+=dr('Role','<span style="color:'+rc(run.role)+'">'+esc(run.role)+'</span>');
  h+=dr('Agent',  esc(run.agent_name||'—'));
  h+=dr('Rig',    esc(run.rig||'town'));
  h+=dr('Started',esc(fmtTs(run.started_at)));
  if(run.ended_at) h+=dr('Ended',esc(fmtTs(run.ended_at)));
  h+=dr('Duration',run.duration_ms?fmtMs(run.duration_ms):run.running?'<span style="color:#56d364">running</span>':'?');
  if(run.cost)     h+=dr('Cost','$'+run.cost.toFixed(4));
  if(run.run_id)   h+=dr('Run ID','<span style="font-size:9px;word-break:break-all">'+esc(run.run_id)+'</span>');
  if(run.session_id) h+=dr('Session','<span style="font-size:9px;word-break:break-all">'+esc(run.session_id)+'</span>');
  h+=dr('Events',(run.events||[]).length);

  h+='<div class="ov-dsect" style="margin-top:8px">Events ('+((run.events||[]).length)+')</div>';
  h+='<div class="ov-evlist">';
  var evs=run.events||[];
  for(var i=0;i<evs.length;i++){
    var ev=evs[i];
    h+='<div class="ov-evitem" style="border-left-color:'+ec(ev.body)+'">'
      +'<span class="ov-evtime">'+fmtTs(ev.timestamp).substring(11,19)+'</span>'
      +'<span class="ov-evbody" style="color:'+ec(ev.body)+'">'+esc(ev.body)+'</span>'
      +'</div>';
  }
  h+='</div>';
  detBody.innerHTML=h;
  detPanel.style.display='flex';
}

function closeDetail(){
  detPanel.style.display='none';
  selectedRun=null;
  drawMain(); drawLabels();
}
window.closeDetail=closeDetail;

function dr(k,v){
  return '<div class="ov-drow"><div class="ov-dkey">'+k+'</div><div class="ov-dval">'+v+'</div></div>';
}

// ── Public filters ────────────────────────────────────────────────────────────
function applyFilters(){
  var params=new URLSearchParams(window.location.search);
  ['rig','role'].forEach(function(k){
    var v=document.getElementById('f-'+k).value;
    v?params.set(k,v):params.delete(k);
  });
  window.history.replaceState(null,'','?'+params.toString());
  applyFiltersLocal();
  buildLanes();
  computeRange();
  vStart=fullStart; vEnd=fullEnd;
  updateSummary();
  resize();
}
window.applyFilters=applyFilters;

function resetFilters(){
  var p=new URLSearchParams(window.location.search);
  ['rig','role'].forEach(function(k){ p.delete(k); });
  window.location.href='?'+p.toString();
}
window.resetFilters=resetFilters;

// ── Summary cards ─────────────────────────────────────────────────────────────
function updateSummary(){
  var s=DATA.summary;
  set('s-runs',   s.run_count||0);
  set('s-rigs',   s.rig_count||0);
  set('s-events', s.event_count||0);
  set('s-cost',   '$'+((s.total_cost||0).toFixed(4)));
  set('s-dur',    s.total_duration||'—');
}
function set(id,v){ var el=document.getElementById(id); if(el) el.textContent=v; }

// ── Utilities ─────────────────────────────────────────────────────────────────
function pad2(n){ return n.toString().padStart(2,'0'); }
function fmtMs(ms){
  if(ms<1000) return Math.round(ms)+'ms';
  if(ms<60000) return (ms/1000).toFixed(1)+'s';
  return Math.floor(ms/60000)+'m'+Math.floor((ms%60000)/1000)+'s';
}
function fmtTs(ts){
  if(!ts) return '—';
  return ts.replace('T',' ').replace('Z','').substring(0,19);
}
function esc(s){
  return String(s).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');
}
function niceInterval(rangeMs,pixW){
  var ivs=[500,1000,2000,5000,10000,15000,30000,60000,120000,300000,600000,1800000,3600000,7200000];
  for(var i=0;i<ivs.length;i++) if(pixW/(rangeMs/ivs[i])>=70) return ivs[i];
  return ivs[ivs.length-1];
}
function rulerLabel(t){
  var d=new Date(t), r=fullEnd-fullStart;
  return pad2(d.getHours())+':'+(r<90000?pad2(d.getMinutes())+':'+pad2(d.getSeconds()):pad2(d.getMinutes()));
}

window.addEventListener('load', init);
})();
</script>
{{end}}
`
