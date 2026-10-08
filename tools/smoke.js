// Smoke test for the add-on features, run against a live WhatsApp Web page over DevTools.
//
//   1. Quit the app, then:  powershell -ExecutionPolicy Bypass -File tools\dev.ps1
//   2. node tools\smoke.js            (env WA_CDP_PORT overrides the auto-discovered port)
//
// It opens one chat that has no unread messages (so no read receipts are sent), types test text into the
// composer and clears it again. It never presses Enter and never sends anything. It prints only test
// strings and counts, not chat content.

const fs = require('fs');
const path = require('path');

// DevTools port: WA_CDP_PORT, else the one dev.ps1 started (written by WebView2 into the profile folder).
function discoverPort() {
  if (process.env.WA_CDP_PORT) return process.env.WA_CDP_PORT;
  const f = path.join(process.env.APPDATA || '', 'WhatsAppDesktopLight', 'UserData', 'EBWebView', 'DevToolsActivePort');
  try { return fs.readFileSync(f, 'utf8').split(/\r?\n/)[0].trim(); } catch (e) { return 9222; }
}
const PORT = discoverPort();
const results = [];
const sleep = ms => new Promise(r => setTimeout(r, ms));

(async () => {
  const list = await (await fetch(`http://127.0.0.1:${PORT}/json`)).json();
  const page = list.find(t => t.type === 'page' && /web\.whatsapp\.com/.test(t.url));
  if (!page) throw new Error('No WhatsApp page found on port ' + PORT);
  const ws = new WebSocket(page.webSocketDebuggerUrl);
  await new Promise(r => (ws.onopen = r));
  let id = 0;
  const pending = {};
  ws.onmessage = e => { const m = JSON.parse(e.data); if (pending[m.id]) pending[m.id](m); };
  const send = (method, params) => new Promise(r => { const i = ++id; pending[i] = r; ws.send(JSON.stringify({ id: i, method, params })); });
  const ev = async expr => {
    const r = await send('Runtime.evaluate', { expression: expr, returnByValue: true, awaitPromise: true });
    if (r.result.exceptionDetails) throw new Error('page error: ' + JSON.stringify(r.result.exceptionDetails.exception || r.result.exceptionDetails));
    return r.result.result.value;
  };
  const key = async (k, code, vk, mods = 0, commands) => {
    for (const t of ['rawKeyDown', 'keyUp'])
      await send('Input.dispatchKeyEvent', { type: t, key: k, code, windowsVirtualKeyCode: vk, modifiers: mods, commands: t === 'rawKeyDown' ? commands : undefined });
  };
  const check = (name, ok, detail = '') => { results.push(ok); console.log((ok ? 'PASS' : 'FAIL') + '  ' + name + (detail ? '  ' + detail : '')); };

  const COMP = `document.querySelector('footer div[contenteditable="true"]')`;
  const composerText = () => ev(`(()=>{var c=${COMP};return c?c.innerText.trim():null})()`);
  const clear = async () => { await ev(`${COMP}.focus()`); await key('a', 'KeyA', 65, 2, ['selectAll']); await key('Backspace', 'Backspace', 8); await sleep(400); };
  const type = async s => { for (const ch of s) { await send('Input.insertText', { text: ch }); await sleep(70); } await sleep(500); };

  // ---- 1. Add-on loaded and selectors resolve
  check('add-on core loaded', (await ev(`typeof window.waSel`)) === 'object');
  const diag = await ev(`JSON.stringify(window.waRunDiagnostics())`).then(JSON.parse);
  const bad = diag.filter(d => d.state === 'bad');
  check('diagnostics: no missing selectors', bad.length === 0, bad.map(d => d.name).join(', '));

  // ---- 2. Unread filter drives the native chip
  const tabs = `[...document.querySelectorAll('[role="tablist"][aria-label="chat-list-filters"] [role="tab"]')]`;
  const selectedTab = () => ev(`${tabs}.filter(t=>t.getAttribute('aria-selected')==='true').map(t=>t.textContent.replace(/[0-9]/g,'').trim()).join()`);
  await ev(`window.toggleUnreadFilter(true)`); await sleep(1200);
  const onTab = await selectedTab();
  await ev(`window.toggleUnreadFilter(false)`); await sleep(1200);
  const offTab = await selectedTab();
  check('unread filter on/off', /belum dibaca|unread/i.test(onTab) && /semua|all/i.test(offTab), `on="${onTab}" off="${offTab}"`);

  // ---- 3. Open a chat without unread messages
  await ev(`(async()=>{ if(document.querySelector('#main')) return; var row=waSel.all('chatRows').find(r=>!r.querySelector('[data-testid="icon-unread-count"]')); if(!row) return; var t=row.querySelector('[data-testid="cell-frame-container"]'); ['mousedown','mouseup','click'].forEach(e=>t.dispatchEvent(new MouseEvent(e,{bubbles:true,cancelable:true,view:window}))); await new Promise(r=>setTimeout(r,2500)); })()`);
  const chatOpen = await ev(`!!document.querySelector('#main')`);
  check('opened a chat', chatOpen);

  // ---- 4. Pins: pin, chip appears, chip opens chat, unpin
  const pinned = await ev(`(async()=>{
    var rows=waSel.all('chatRows'); var btn=rows[1].querySelector('.wa-pin-btn'); if(!btn) return 'no-toggle';
    btn.dispatchEvent(new MouseEvent('click',{bubbles:true,cancelable:true})); await new Promise(r=>setTimeout(r,900));
    var chips=document.querySelectorAll('#wa-pin-bar .wa-pin-chip').length;
    var b2=waSel.all('chatRows').find(r=>r.querySelector('.wa-pin-btn.pinned')); if(b2) b2.querySelector('.wa-pin-btn').dispatchEvent(new MouseEvent('click',{bubbles:true,cancelable:true}));
    await new Promise(r=>setTimeout(r,700));
    return chips+'/'+document.querySelectorAll('#wa-pin-bar .wa-pin-chip').length })()`);
  check('pin then unpin', pinned === '1/0', pinned);

  // ---- 5. Radar: fires a toast when the header turns "online"
  const radar = await ev(`(()=>{
    window.waRadar.config.enabled=true; window.waRadar.config.names=''; window.waRadar.tick();
    var h=waSel.one('chatHeader'); if(!h) return 'no-header';
    var f=document.createElement('span'); f.textContent='online'; h.appendChild(f);
    var fired=0, orig=window.showAddonToast; window.showAddonToast=function(){fired++};
    window.waRadar.tick(); f.remove(); window.showAddonToast=orig; window.waRadar.config.enabled=false; return fired })()`);
  check('radar toast', radar === 1, String(radar));

  // ---- 6. Quick replies (needs the composer). Uses the "/thanks" -> "terima kasih" reply if it exists.
  const replies = JSON.parse((await ev(`localStorage.getItem('wa_quick_replies')`)) || '[]');
  const thanks = replies.find(r => r.key === '/thanks');
  if (!chatOpen || !(await ev(`!!${COMP}`))) {
    check('quick replies', false, 'no composer');
  } else if (!thanks) {
    console.log('SKIP  quick replies: no "/thanks" reply defined');
  } else {
    const want = thanks.text;
    await clear(); await type('/thanks');
    check('quick reply popup shows', await ev(`(()=>{var p=document.querySelector('#wa-qr-popup');return !!p&&getComputedStyle(p).display!=='none'})()`));
    await key('Tab', 'Tab', 9); await sleep(900);
    check('quick reply via Tab', (await composerText()) === want, JSON.stringify(await composerText()));

    await clear(); await type('/thanks');
    await ev(`document.querySelector('#wa-qr-popup .wa-qr-item').dispatchEvent(new MouseEvent('mousedown',{bubbles:true,cancelable:true}))`); await sleep(900);
    check('quick reply via popup click', (await composerText()) === want);

    await clear(); await type('/thanks'); await key(' ', 'Space', 32); await sleep(900);
    check('quick reply via Space', (await composerText()) === want);

    await clear(); await type('halo kak /thanks'); await key('Tab', 'Tab', 9); await sleep(900);
    check('quick reply mid-sentence', (await composerText()) === 'halo kak ' + want);

    const multi = replies.find(r => r.text.includes('\n'));
    if (multi) {
      await clear(); await type(multi.key); await key('Tab', 'Tab', 9); await sleep(900);
      const paras = await ev(`${COMP}.querySelectorAll('p').length`);
      check('multi-line reply keeps line breaks', paras === multi.text.split('\n').length, `paragraphs=${paras}`);
    }
    await clear();
    check('composer left empty', (await composerText()) === '');
  }

  // ---- 7. Per-chat blur toggle
  const blur = await ev(`(async()=>{
    var wait=ms=>new Promise(r=>setTimeout(r,ms));
    var eye=waSel.all('chatRows')[1].querySelector('.wa-blur-toggle'); if(!eye) return 'no-toggle';
    eye.dispatchEvent(new MouseEvent('click',{bubbles:true,cancelable:true})); await wait(900);
    var row=waSel.all('chatRows')[1];
    var on=row.getAttribute('data-wa-blur');
    var filt=getComputedStyle(row.querySelector('[data-testid="cell-frame-title"]')).filter;
    row.querySelector('.wa-blur-toggle').dispatchEvent(new MouseEvent('click',{bubbles:true,cancelable:true})); await wait(700);
    return on+'|'+(filt!=='none')+'|'+waSel.all('chatRows')[1].getAttribute('data-wa-blur') })()`);
  check('per-chat blur on/off', blur === '1|true|0', blur);

  // ---- 8. Quick reply variables
  const vars = await ev(`window.waExpandVars('Halo {depan}, hari {hari} jam {jam} {unknown}')`);
  check('quick reply variables', !/\{(depan|hari|jam)\}/.test(vars) && vars.endsWith('{unknown}') && /\d\d:\d\d/.test(vars));

  // ---- 9. Notes: create two, search, restore previous notes afterwards
  const notes = await ev(`(async()=>{
    var wait=ms=>new Promise(r=>setTimeout(r,ms));
    var before=localStorage.getItem('wa_notes');
    window.toggleScratchpad(); await wait(500);
    var d=document.getElementById('wa-scratchpad-drawer'); if(!d) return 'no-drawer';
    var open=d.style.right==='0px';
    document.getElementById('wa-note-new').click(); await wait(100);
    var ta=document.getElementById('wa-sp-textarea'); ta.value='smoke-note-alpha'; ta.dispatchEvent(new Event('input',{bubbles:true}));
    document.getElementById('wa-note-new').click(); await wait(100);
    ta=document.getElementById('wa-sp-textarea'); ta.value='smoke-note-beta'; ta.dispatchEvent(new Event('input',{bubbles:true}));
    var s=document.getElementById('wa-note-search'); s.value='alpha'; s.dispatchEvent(new Event('input',{bubbles:true}));
    var filtered=document.querySelectorAll('#wa-note-list .wa-note-item').length;
    s.value=''; s.dispatchEvent(new Event('input',{bubbles:true}));
    var total=document.querySelectorAll('#wa-note-list .wa-note-item').length;
    if(before===null) localStorage.removeItem('wa_notes'); else localStorage.setItem('wa_notes',before);
    window.toggleScratchpad(); await wait(300); d.remove();
    return open+'|'+filtered+'|'+(total>=2) })()`);
  check('notes: create, search, list', notes === 'true|1|true', notes);

  // ---- 10. Chat export builds lines from the open chat
  const exportLines = await ev(`window.waBuildChatExport().length`);
  check('chat export builds lines', chatOpen && exportLines > 0, 'lines=' + exportLines);

  // ---- 11. Lock when hidden
  const lock = await ev(`(async()=>{
    var prev=localStorage.getItem('wa_lock_on_hide');
    localStorage.setItem('wa_lock_on_hide','true'); window.waOnHidden(); await new Promise(r=>setTimeout(r,500));
    var o=document.getElementById('wa-lock-overlay'); var shown=!!o&&getComputedStyle(o).display!=='none';
    if(o) o.style.display='none'; try{localStorage.setItem('wa_is_locked','false')}catch(e){}
    if(prev===null) localStorage.removeItem('wa_lock_on_hide'); else localStorage.setItem('wa_lock_on_hide',prev);
    return shown })()`);
  check('lock when hidden', lock === true, String(lock));

  const failed = results.filter(r => !r).length;
  console.log(failed ? `\n${failed} FAILED` : '\nALL PASS');
  process.exit(failed ? 1 : 0);
})().catch(e => { console.error('ERROR', e.message); process.exit(2); });
