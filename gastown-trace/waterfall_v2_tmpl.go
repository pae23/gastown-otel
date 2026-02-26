package main

const tmplWaterfallV2 = `
{{define "content"}}
<style>
/* ── Waterfall V2 — event table layout ───────────────────────────────────── */
.wf2-toolbar{display:flex;gap:8px;align-items:center;flex-wrap:wrap;padding:4px 0 8px}
.wf2-toolbar select,.wf2-toolbar input[type=text]{background:#161b22;border:1px solid #30363d;color:#c9d1d9;padding:4px 8px;border-radius:4px;font-size:12px;font-family:monospace}
.wf2-toolbar label{font-size:11px;color:#8b949e}
.wf2-toolbar button{background:#1f6feb;border:none;color:#fff;padding:4px 12px;border-radius:4px;cursor:pointer;font-size:12px}
.wf2-toolbar button.sec{background:#21262d;border:1px solid #30363d;color:#c9d1d9}
.wf2-cards{display:flex;gap:8px;flex-wrap:wrap;margin-bottom:8px}
.wf2-card{background:#161b22;border:1px solid #30363d;border-radius:6px;padding:10px 16px;min-width:90px}
.wf2-card-val{font-size:22px;font-weight:700;color:#c9d1d9;font-family:monospace}
.wf2-card-lbl{font-size:10px;color:#8b949e;text-transform:uppercase;letter-spacing:.5px}
.wf2-instance{font-size:11px;color:#58a6ff;margin-bottom:6px}
/* Overview */
#wf-ov-wrap{border:1px solid #30363d;border-radius:4px;background:#0d1117;margin-bottom:6px;cursor:crosshair;overflow:hidden;user-select:none;position:relative}
#wf-ov{display:block}
#wf-tooltip{position:fixed;background:#1c2128;border:1px solid #30363d;padding:8px 10px;border-radius:4px;font-size:11px;color:#c9d1d9;pointer-events:none;display:none;z-index:200;max-width:260px;line-height:1.6;box-shadow:0 4px 12px rgba(0,0,0,.5)}
/* Main flex area */
#wf-main{display:flex;gap:8px;align-items:flex-start}
/* Table */
.wf-table-wrap{flex:1;min-width:0;overflow:auto;max-height:65vh;border:1px solid #30363d;border-radius:4px;background:#0d1117}
.wf-table{width:100%;border-collapse:collapse;font-size:11px;table-layout:fixed}
.wf-table th{background:#161b22;color:#8b949e;text-align:left;padding:6px 8px;border-bottom:1px solid #30363d;position:sticky;top:0;z-index:1;font-weight:600;white-space:nowrap;user-select:none}
.wf-table td{padding:3px 8px;border-bottom:1px solid #161b22;vertical-align:middle;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.wf-table tr.wf-row:hover td{background:#1c2128;cursor:pointer}
.wf-table tr.wf-sel td{background:#1f6feb1a!important}
.wf-table tr.wf-sel td:first-child{border-left:2px solid #58a6ff}
.err-row td{border-left:2px solid #ef4444}
.wf-bar-outer{position:relative;height:10px;background:#0d1117;overflow:hidden}
.wf-bar-inner{position:absolute;top:0;height:10px;border-radius:1px;min-width:2px;opacity:0.85}
.wf-row-count{font-size:10px;color:#8b949e;background:#161b22;border:1px solid #30363d;padding:2px 8px;border-radius:4px}
/* Detail panel */
#wf-detail{display:none;flex-direction:column;width:380px;min-width:300px;flex-shrink:0;border:1px solid #30363d;border-radius:4px;max-height:65vh;overflow:hidden}
#wf-detail-content{display:flex;flex-direction:column;flex:1;min-height:0;overflow:hidden}
.wf2-dhdr{display:flex;justify-content:space-between;align-items:center;padding:8px 12px;background:#161b22;border-bottom:1px solid #30363d;font-size:12px;flex-wrap:wrap;gap:6px;flex-shrink:0}
.wf2-dhdr button{background:#21262d;border:1px solid #30363d;color:#c9d1d9;padding:2px 8px;border-radius:4px;cursor:pointer;font-size:11px}
.wf2-devents{overflow-y:auto;flex:1;min-height:0;padding:0 0 8px}
.wf2-devents table{width:100%;border-collapse:collapse;font-size:11px}
.wf2-devents th{color:#8b949e;text-align:left;padding:4px 8px;font-weight:500;white-space:nowrap;vertical-align:top;width:110px;border-bottom:1px solid #1c2128}
.wf2-devents td{padding:3px 8px;border-bottom:1px solid #161b22;vertical-align:top;word-break:break-all}
.wf2-devents tr:hover td{background:#1c2128}
.wf2-pre{background:#161b22;padding:6px 8px;border-radius:4px;font-size:9px;color:#c9d1d9;margin:2px 0;white-space:pre-wrap;word-break:break-all;max-height:180px;overflow-y:auto;overflow-x:auto}
.wf2-section-hdr{color:#58a6ff;padding:8px 8px 3px;font-size:9px;font-weight:700;letter-spacing:.8px;text-transform:uppercase;border-bottom:1px solid #1c2128;background:#0d1117;position:sticky;top:0}
/* Event badges */
.ev-b{padding:1px 5px;border-radius:3px;font-size:9px;font-family:monospace;white-space:nowrap}
.ev-agent{background:#10b98122;color:#10b981}.ev-api{background:#f59e0b22;color:#f59e0b}
.ev-tool{background:#3b82f622;color:#3b82f6}.ev-bd{background:#ef444422;color:#ef4444}
.ev-sling{background:#f59e0b22;color:#f59e0b}.ev-mail{background:#eab30822;color:#eab308}
.ev-nudge{background:#8b5cf622;color:#8b5cf6}.ev-done{background:#56d36422;color:#56d364}
.ev-prime{background:#3b82f622;color:#3b82f6}.ev-prompt{background:#06b6d422;color:#06b6d4}
.ev-sess{background:#6b728022;color:#6b7280}.ev-inst{background:#8b5cf622;color:#8b5cf6}
.ev-other{background:#30363d;color:#8b949e}
</style>

<div id="wf2-page">

{{if .Instance}}<div class="wf2-instance">instance: <b>{{.Instance}}</b>{{if .TownRoot}} &nbsp;·&nbsp; {{.TownRoot}}{{end}}</div>{{end}}

<div class="wf2-toolbar">
  <label>Rig</label>
  <select id="f-rig"><option value="">All rigs</option></select>
  <label>Role</label>
  <select id="f-role">
    <option value="">All roles</option>
    <option>mayor</option><option>deacon</option><option>witness</option>
    <option>refinery</option><option>polecat</option><option>dog</option>
    <option>boot</option><option>crew</option>
  </select>
  <label>Type</label>
  <select id="f-type"><option value="">All types</option></select>
  <label>Search</label>
  <input type="text" id="f-search" placeholder="agent, bead ID…" style="width:140px">
  <button onclick="applyFilters()">Apply</button>
  <button class="sec" onclick="resetFilters()">Reset</button>
  <span style="flex:1"></span>
  <span class="wf-row-count" id="wf-row-count">—</span>
  <span style="font-size:11px;color:#8b949e" id="wf2-winlabel">window: {{.Window}}</span>
</div>

<div class="wf2-cards">
  <div class="wf2-card"><div class="wf2-card-val" id="wf2-s-runs">—</div><div class="wf2-card-lbl">Runs</div></div>
  <div class="wf2-card"><div class="wf2-card-val" id="wf2-s-rigs">—</div><div class="wf2-card-lbl">Rigs</div></div>
  <div class="wf2-card"><div class="wf2-card-val" id="wf2-s-events">—</div><div class="wf2-card-lbl">Events</div></div>
  <div class="wf2-card"><div class="wf2-card-val" id="wf2-s-beads">—</div><div class="wf2-card-lbl">Beads</div></div>
  <div class="wf2-card"><div class="wf2-card-val" id="wf2-s-cost">—</div><div class="wf2-card-lbl">Cost</div></div>
  <div class="wf2-card"><div class="wf2-card-val" id="wf2-s-dur">—</div><div class="wf2-card-lbl">Total duration</div></div>
</div>

<div id="wf-ov-wrap"><canvas id="wf-ov"></canvas></div>
<div id="wf-tooltip"></div>

<div id="wf-main">
  <div class="wf-table-wrap">
    <table class="wf-table">
      <thead>
        <tr>
          <th style="width:64px">Time</th>
          <th style="width:140px">Type</th>
          <th style="width:110px">Agent</th>
          <th style="width:72px">Rig</th>
          <th>Detail</th>
          <th style="width:54px">Dur</th>
          <th style="width:110px">Waterfall</th>
        </tr>
      </thead>
      <tbody id="wf-tbody"></tbody>
    </table>
  </div>
  <div id="wf-detail">
    <div id="wf-detail-content"></div>
  </div>
</div>

</div><!-- #wf2-page -->

<script>
(function(){
// ── Data ──────────────────────────────────────────────────────────────────────
var DATA = {{.JSONData}};

// ── Constants ─────────────────────────────────────────────────────────────────
var ROLE_CLR = {
  mayor:'#f59e0b', deacon:'#8b5cf6', witness:'#3b82f6',
  refinery:'#10b981', polecat:'#ef4444', dog:'#f97316',
  boot:'#6b7280', crew:'#06b6d4'
};
function rc(role){ return ROLE_CLR[role]||'#9ca3af'; }
var MAX_ROWS = 2000;

// ── State ─────────────────────────────────────────────────────────────────────
var vStart=0, vEnd=0, fullStart=0, fullEnd=0;
var allEvents=[], selectedRow=null, filteredData=null;
var ovCanvas, ovCtx, detPanel, detContent, tip;
var ovDrag=false, ovDragMode=null, ovDragX0=0, ovDragVS0=0, ovDragVE0=0;

// ── Init ──────────────────────────────────────────────────────────────────────
function init(){
  ovCanvas   = document.getElementById('wf-ov');
  ovCtx      = ovCanvas.getContext('2d');
  detPanel   = document.getElementById('wf-detail');
  detContent = document.getElementById('wf-detail-content');
  tip        = document.getElementById('wf-tooltip');

  var ovWrap = document.getElementById('wf-ov-wrap');
  ovCanvas.width  = ovWrap.clientWidth||900;
  ovCanvas.height = 56;
  ovWrap.style.height = '56px';

  filteredData = DATA;

  // Populate rig filter
  var selRig = document.getElementById('f-rig');
  for(var i=0;i<DATA.rigs.length;i++){
    var o=document.createElement('option');
    o.value=o.textContent=DATA.rigs[i].name;
    selRig.appendChild(o);
  }

  // Restore URL params
  var params=new URLSearchParams(window.location.search);
  if(params.get('rig'))  document.getElementById('f-rig').value  = params.get('rig');
  if(params.get('role')) document.getElementById('f-role').value = params.get('role');
  if(params.get('type')) document.getElementById('f-type').value = params.get('type');
  if(params.get('q'))    document.getElementById('f-search').value = params.get('q');

  applyFiltersLocal();
  buildAllEvents();
  computeFullRange();
  vStart=fullStart; vEnd=fullEnd;
  populateTypeFilter();
  updateSummary();
  renderTable();
  drawOverview();

  ovCanvas.addEventListener('mousedown', onOvDown);
  ovCanvas.addEventListener('mousemove', onOvMove);
  ovCanvas.addEventListener('mouseleave', function(){ ovCanvas.style.cursor='crosshair'; });
  window.addEventListener('mousemove', onWinMove);
  window.addEventListener('mouseup',   onWinUp);
  window.addEventListener('resize',    onResize);
  window.addEventListener('keydown',   function(e){ if(e.key==='Escape') closeDetail(); });
}

// ── Data helpers ──────────────────────────────────────────────────────────────
function applyFiltersLocal(){
  var rigF  = document.getElementById('f-rig').value;
  var roleF = document.getElementById('f-role').value;
  var qF    = document.getElementById('f-search').value.toLowerCase();
  var rigs  = DATA.rigs.map(function(rig){
    if(rigF && rig.name !== rigF) return null;
    var runs = rig.runs.filter(function(run){
      if(roleF && run.role !== roleF) return false;
      if(qF){
        var hay=(run.run_id+run.agent_name+run.role+run.session_id+run.rig).toLowerCase();
        if(hay.indexOf(qF)<0) return false;
      }
      return true;
    });
    if(runs.length===0) return null;
    return {name:rig.name, collapsed:rig.collapsed, runs:runs};
  }).filter(Boolean);
  filteredData = {
    instance:DATA.instance, town_root:DATA.town_root,
    window:DATA.window, summary:DATA.summary,
    rigs:rigs, communications:DATA.communications, beads:DATA.beads
  };
}

function buildAllEvents(){
  allEvents = [];
  for(var i=0;i<filteredData.rigs.length;i++){
    var rig=filteredData.rigs[i];
    for(var j=0;j<rig.runs.length;j++){
      var run=rig.runs[j];
      var evs=run.events||[];
      for(var k=0;k<evs.length;k++){
        allEvents.push({ev:evs[k], run:run, rigName:rig.name});
      }
    }
  }
  allEvents.sort(function(a,b){ return a.ev.timestamp<b.ev.timestamp?-1:1; });
}

function computeFullRange(){
  var mn=Infinity, mx=-Infinity;
  for(var i=0;i<allEvents.length;i++){
    var t=new Date(allEvents[i].ev.timestamp).getTime();
    if(t<mn) mn=t; if(t>mx) mx=t;
  }
  if(!isFinite(mn)){ mn=Date.now()-3600000; mx=Date.now(); }
  var pad=(mx-mn)*0.04||5000;
  fullStart=mn-pad; fullEnd=mx+pad;
}

function populateTypeFilter(){
  var types={};
  for(var i=0;i<allEvents.length;i++) types[allEvents[i].ev.body]=1;
  var sel=document.getElementById('f-type');
  while(sel.options.length>1) sel.remove(1);
  Object.keys(types).sort().forEach(function(k){
    var o=document.createElement('option'); o.value=o.textContent=k; sel.appendChild(o);
  });
}

// ── Table rendering ───────────────────────────────────────────────────────────
function renderTable(){
  var fType = document.getElementById('f-type').value;
  var fQ    = document.getElementById('f-search').value.toLowerCase();
  var range = vEnd-vStart;

  var rows = allEvents.filter(function(row){
    var t=new Date(row.ev.timestamp).getTime();
    if(t<vStart||t>vEnd) return false;
    if(fType && row.ev.body !== fType) return false;
    if(fQ){
      var hay=(row.run.agent_name+row.run.role+row.rigName+row.ev.body+JSON.stringify(row.ev.attrs||{})).toLowerCase();
      if(hay.indexOf(fQ)<0) return false;
    }
    return true;
  });

  var total=rows.length, capped=rows.length>MAX_ROWS;
  if(capped) rows=rows.slice(0,MAX_ROWS);

  var html='';
  for(var i=0;i<rows.length;i++){
    var row=rows[i], ev=row.ev, run=row.run;
    var d=new Date(ev.timestamp);
    var ts=pad2(d.getHours())+':'+pad2(d.getMinutes())+':'+pad2(d.getSeconds());
    var dur=ev.attrs&&ev.attrs.duration_ms?fmtMs(parseFloat(ev.attrs.duration_ms)):'';
    var isSel=(selectedRow&&selectedRow.ev.id===ev.id);
    var agentLabel=(run.agent_name||run.role||'').substring(0,14);
    html+='<tr class="wf-row'+(ev.severity==='error'?' err-row':'')+(isSel?' wf-sel':'')+'" data-i="'+i+'">'
      +'<td class="mono dim" style="font-size:10px">'+ts+'</td>'
      +'<td>'+evBadge(ev.body)+'</td>'
      +'<td class="mono" title="'+esc(run.agent_name||run.role||'')+'"><span style="color:'+rc(run.role)+'">'+esc(agentLabel)+'</span></td>'
      +'<td class="mono dim" style="font-size:10px">'+esc(row.rigName||'')+'</td>'
      +'<td class="mono">'+evDetail(ev)+'</td>'
      +'<td class="mono dim" style="font-size:10px">'+dur+'</td>'
      +'<td>'+makeWfBar(ev,range)+'</td>'
      +'</tr>';
  }

  var capturedRows=rows;
  var tbody=document.getElementById('wf-tbody');
  tbody.innerHTML=html;
  var trs=tbody.querySelectorAll('tr');
  for(var j=0;j<trs.length;j++){
    (function(tr,row){ tr.addEventListener('click',function(){ showDetail(row); }); })(trs[j],capturedRows[j]);
  }

  var countEl=document.getElementById('wf-row-count');
  if(countEl) countEl.textContent=(capped?MAX_ROWS+' / ':'')+total+' events';
}

function makeWfBar(ev, range){
  var t=new Date(ev.timestamp).getTime();
  var x=((t-vStart)/range)*100;
  if(x<0||x>100) return '';
  var dur=ev.attrs&&ev.attrs.duration_ms?parseFloat(ev.attrs.duration_ms):0;
  var w=dur>0?Math.max(2,(dur/range)*100):2;
  var color='#f59e0b';
  if(ev.body==='claude_code.tool_result') color=(ev.attrs&&ev.attrs.success!=='false')?'#10b981':'#ef4444';
  else if(ev.body==='agent.event')        color='#8b5cf6';
  else if(ev.body==='bd.call')            color='#3b82f6';
  else if(ev.body==='sling'||ev.body==='mail') color='#eab308';
  else if(ev.body==='done'||ev.body==='session.stop') color='#56d364';
  else if(ev.body==='agent.instantiate'||ev.body==='session.start') color='#06b6d4';
  return '<div class="wf-bar-outer"><div class="wf-bar-inner" style="left:'+x.toFixed(1)+'%;width:'+w.toFixed(1)+'%;background:'+color+'"></div></div>';
}

// ── Event helpers ─────────────────────────────────────────────────────────────
function evBadge(body){
  var m={
    'agent.event':'ev-agent','claude_code.api_request':'ev-api',
    'claude_code.tool_result':'ev-tool','bd.call':'ev-bd',
    'sling':'ev-sling','mail':'ev-mail','nudge':'ev-nudge','done':'ev-done',
    'prime':'ev-prime','prompt.send':'ev-prompt',
    'session.start':'ev-sess','session.stop':'ev-sess',
    'agent.instantiate':'ev-inst'
  };
  return '<span class="ev-b '+(m[body]||'ev-other')+'">'+esc(body)+'</span>';
}

function evDetail(ev){
  var a=ev.attrs||{};
  if(ev.body==='agent.event')             return esc((a.content||'').substring(0,120));
  if(ev.body==='bd.call')                 return esc(((a.subcommand||'')+' '+(a.args||'')).substring(0,100));
  if(ev.body==='claude_code.api_request') return esc((a.model||'')+' in:'+(a.input_tokens||0)+' out:'+(a.output_tokens||0)+' $'+parseFloat(a.cost_usd||0).toFixed(4));
  if(ev.body==='claude_code.tool_result') return esc((a.tool_name||'')+' '+(a.success!=='false'?'✓':'✗')+' '+fmtMs(parseFloat(a.duration_ms||0)));
  if(ev.body==='mail')                    return esc(((a.operation||'')+' '+(a['msg.from']||'')+' → '+(a['msg.to']||'')+': '+(a['msg.subject']||'')).substring(0,100));
  if(ev.body==='prompt.send')             return esc((a.keys_len||0)+' bytes'+(a.keys?' — '+a.keys.substring(0,60):''));
  if(ev.body==='done')                    return esc(a.exit_type||'');
  if(ev.body==='prime')                   return esc((a.formula||'').substring(0,80));
  var keys=Object.keys(a).filter(function(k){return k!=='run.id'&&!k.startsWith('gt.');}).slice(0,5);
  return esc(keys.map(function(k){return k+'='+a[k];}).join(' '));
}

// ── Detail panel ──────────────────────────────────────────────────────────────
function showDetail(row){
  selectedRow=row;
  var ev=row.ev, run=row.run, attrs=ev.attrs||{};

  var html='<div class="wf2-dhdr">'
    +'<span>'+evBadge(ev.body)
    +' &nbsp;<span style="color:'+rc(run.role)+';font-size:11px">'+esc(run.agent_name||run.role||'')+'</span></span>'
    +'<button onclick="closeDetail()">✕</button>'
    +'</div>';
  html+='<div class="wf2-devents">';

  // Context
  html+='<div class="wf2-section-hdr">Context</div>';
  html+='<table>';
  html+='<tr><th>Time</th><td class="mono" style="font-size:10px">'+esc(ev.timestamp)+'</td></tr>';
  html+='<tr><th>Agent</th><td class="mono">'+esc(run.agent_name||'—')+'</td></tr>';
  html+='<tr><th>Role</th><td><span style="color:'+rc(run.role)+'">'+esc(run.role||'—')+'</span></td></tr>';
  html+='<tr><th>Rig</th><td class="mono">'+esc(row.rigName||'—')+'</td></tr>';
  html+='<tr><th>Run ID</th><td class="mono" style="font-size:9px;word-break:break-all">'+esc(run.run_id||'—')+'</td></tr>';
  if(run.session_id) html+='<tr><th>Session</th><td class="mono" style="font-size:9px;word-break:break-all">'+esc(run.session_id)+'</td></tr>';
  if(ev.severity==='error') html+='<tr><th>Severity</th><td style="color:#ef4444">error</td></tr>';
  html+='</table>';

  // OTel attributes
  var keys=Object.keys(attrs).sort();
  if(keys.length>0){
    html+='<div class="wf2-section-hdr">OTel Attributes</div>';
    html+='<table>';
    for(var i=0;i<keys.length;i++){
      var k=keys[i], v=String(attrs[k]);
      html+='<tr><th>'+esc(k)+'</th>';
      if(v.length>100||v.indexOf('\n')>=0){
        html+='<td><pre class="wf2-pre">'+esc(v.substring(0,4000))+'</pre></td>';
      } else {
        html+='<td class="mono" style="font-size:10px">'+esc(v)+'</td>';
      }
      html+='</tr>';
    }
    html+='</table>';
  }
  html+='</div>';

  detContent.innerHTML=html;
  detPanel.style.display='flex';
  renderTable();
}

function closeDetail(){
  detPanel.style.display='none';
  selectedRow=null;
  renderTable();
}
window.closeDetail=closeDetail;

// ── Overview canvas ───────────────────────────────────────────────────────────
function ovX(t){ return ((t-fullStart)/(fullEnd-fullStart))*ovCanvas.width; }
function ovT(x){ return fullStart+(x/ovCanvas.width)*(fullEnd-fullStart); }

function drawOverview(){
  var W=ovCanvas.width, H=ovCanvas.height;
  if(!W||!H) return;
  ovCtx.fillStyle='#0d1117';
  ovCtx.fillRect(0,0,W,H);

  // All runs as micro-bars (from full DATA for complete picture)
  var allRuns=[];
  for(var i=0;i<DATA.rigs.length;i++)
    for(var j=0;j<DATA.rigs[i].runs.length;j++)
      allRuns.push(DATA.rigs[i].runs[j]);

  var barZone=H-14;
  var laneH=allRuns.length>0?Math.max(1.5,Math.min(8,barZone/allRuns.length)):4;
  for(var k=0;k<allRuns.length;k++){
    var run=allRuns[k];
    var st=new Date(run.started_at).getTime();
    var et=run.ended_at?new Date(run.ended_at).getTime():fullEnd;
    var x1=ovX(st), x2=Math.max(x1+1,ovX(et));
    var y=2+k*laneH;
    if(y>barZone) break;
    ovCtx.fillStyle=rc(run.role)+'99';
    ovCtx.fillRect(x1,y,x2-x1,Math.max(1,laneH-0.5));
  }

  // Dim outside selection
  var sx1=ovX(vStart), sx2=ovX(vEnd);
  ovCtx.fillStyle='rgba(0,0,0,0.55)';
  if(sx1>0) ovCtx.fillRect(0,0,Math.max(0,sx1),H);
  if(sx2<W) ovCtx.fillRect(sx2,0,W-sx2,H);

  // Selection border + handles
  ovCtx.strokeStyle='#58a6ff'; ovCtx.lineWidth=1;
  ovCtx.strokeRect(sx1+0.5,0.5,Math.max(4,sx2-sx1)-1,H-1);
  ovCtx.fillStyle='#58a6ff';
  ovCtx.fillRect(sx1-1,0,3,H);
  ovCtx.fillRect(sx2-2,0,3,H);

  // Time ruler at bottom
  var iv=niceInterval(fullEnd-fullStart,W);
  var first=Math.ceil(fullStart/iv)*iv;
  ovCtx.font='8px monospace'; ovCtx.textAlign='center';
  for(var t=first;t<=fullEnd;t+=iv){
    var x=ovX(t);
    ovCtx.fillStyle='#30363d'; ovCtx.fillRect(x,H-14,1,4);
    ovCtx.fillStyle='#8b949e'; ovCtx.fillText(rulerLabel(t),x,H-1);
  }
}

function onOvMove(e){
  var r=ovCanvas.getBoundingClientRect();
  var x=e.clientX-r.left;
  var sx1=ovX(vStart), sx2=ovX(vEnd);
  if(Math.abs(x-sx1)<8||Math.abs(x-sx2)<8) ovCanvas.style.cursor='ew-resize';
  else if(x>sx1&&x<sx2)                     ovCanvas.style.cursor='grab';
  else                                       ovCanvas.style.cursor='crosshair';
}

function onOvDown(e){
  var r=ovCanvas.getBoundingClientRect();
  var x=e.clientX-r.left;
  var sx1=ovX(vStart), sx2=ovX(vEnd);
  if(Math.abs(x-sx1)<8){
    ovDragMode='left';
  } else if(Math.abs(x-sx2)<8){
    ovDragMode='right';
  } else if(x>sx1&&x<sx2){
    ovDragMode='pan';
  } else {
    var t=ovT(x), half=(vEnd-vStart)/2;
    vStart=t-half; vEnd=t+half;
    drawOverview(); renderTable(); return;
  }
  ovDrag=true; ovDragX0=x; ovDragVS0=vStart; ovDragVE0=vEnd;
  ovCanvas.style.cursor='grabbing';
  e.preventDefault();
}

function onWinMove(e){
  if(!ovDrag) return;
  var r=ovCanvas.getBoundingClientRect();
  var x=e.clientX-r.left;
  var dt=(x-ovDragX0)/ovCanvas.width*(fullEnd-fullStart);
  if(ovDragMode==='pan'){
    vStart=ovDragVS0+dt; vEnd=ovDragVE0+dt;
  } else if(ovDragMode==='left'){
    vStart=Math.min(ovDragVS0+dt, ovDragVE0-1000);
  } else {
    vEnd=Math.max(ovDragVE0+dt, ovDragVS0+1000);
  }
  drawOverview(); renderTable();
}

function onWinUp(){ if(ovDrag){ ovDrag=false; ovCanvas.style.cursor='crosshair'; } }

function onResize(){
  var ovWrap=document.getElementById('wf-ov-wrap');
  if(ovWrap) ovCanvas.width=ovWrap.clientWidth||900;
  drawOverview();
}

// ── Filters ───────────────────────────────────────────────────────────────────
function applyFilters(){
  var params=new URLSearchParams(window.location.search);
  ['rig','role','type','q'].forEach(function(k){
    var v=document.getElementById('f-'+( k==='q'?'search':k)).value;
    v?params.set(k,v):params.delete(k);
  });
  window.history.replaceState(null,'','?'+params.toString());
  applyFiltersLocal();
  buildAllEvents();
  computeFullRange();
  vStart=fullStart; vEnd=fullEnd;
  populateTypeFilter();
  renderTable();
  drawOverview();
}
window.applyFilters=applyFilters;

function resetFilters(){
  var params=new URLSearchParams(window.location.search);
  ['rig','role','type','q'].forEach(function(k){ params.delete(k); });
  window.location.href='?'+params.toString();
}
window.resetFilters=resetFilters;

// ── Summary ───────────────────────────────────────────────────────────────────
function updateSummary(){
  var s=DATA.summary;
  set('wf2-s-runs',  s.run_count||0);
  set('wf2-s-rigs',  s.rig_count||0);
  set('wf2-s-events',s.event_count||0);
  set('wf2-s-beads', s.bead_count||0);
  set('wf2-s-cost',  '$'+((s.total_cost||0).toFixed(4)));
  set('wf2-s-dur',   s.total_duration||'—');
}
function set(id,v){ var el=document.getElementById(id); if(el) el.textContent=v; }

// ── Utilities ─────────────────────────────────────────────────────────────────
function pad2(n){ return n.toString().padStart(2,'0'); }
function fmtMs(ms){
  if(ms<1000) return Math.round(ms)+'ms';
  if(ms<60000) return (ms/1000).toFixed(1)+'s';
  return Math.floor(ms/60000)+'m'+Math.floor((ms%60000)/1000)+'s';
}
function esc(s){
  return String(s).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');
}
function niceInterval(rangeMs,pixW){
  var ivs=[500,1000,2000,5000,10000,15000,30000,60000,120000,300000,600000,1800000,3600000,7200000];
  for(var i=0;i<ivs.length;i++){
    if(pixW/(rangeMs/ivs[i])>=70) return ivs[i];
  }
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
