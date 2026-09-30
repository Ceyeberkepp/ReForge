import { FormEvent, useEffect, useState } from "react";
import { Boxes, Building2, Computer, HardDrive, LayoutDashboard, Package, Rocket, Settings, ShieldCheck, Plus, Trash2, RefreshCw, Play, CheckCircle2, XCircle, Loader2, Network, Database, ClipboardList, KeyRound, Search, Bell, HelpCircle, Sun, Moon, Monitor, Disc3, Copy, ListChecks, Wrench, BarChart3, UserCog, Apple, LogOut, Activity, Server } from "lucide-react";

const API = import.meta.env.VITE_API_URL ?? "";
type Stats={images:number;departments:number;software:number;hosts:number;queued_deployments:number;running_deployments?:number};
type Img={id:string;name:string;os_name:string;os_version:string;architecture:string;image_version:string;image_path?:string|null;sysprep_ready:boolean;checksum?:string|null;notes:string};
type Sw={id:string;name:string;version:string;installer_path:string;silent_install:string;detection_rule:string;uninstall_command:string;install_order:number;required_by_default:boolean};
type Dept={id:string;name:string;code:string;image_id?:string|null;computer_name_pattern:string;ad_ou:string;required_software_ids:string[];optional_software_ids:string[];printers:string[];post_install_scripts:string[];config:any};
type Host={id:string;hostname:string;mac_address:string;serial_number:string;manufacturer:string;model:string;department_id?:string|null;last_seen?:string|null};
type Job={id:string;host_id:string;image_id:string;department_id?:string|null;optional_software:string[];status:string;progress:number;current_step:string;created_at:string};
type Dir={id?:string;domain:string;domain_controller:string;protocol:string;port:number;base_dn:string;service_account:string;default_computer_ou:string};

const navGroups=[
  {label:"",items:[["Dashboard",LayoutDashboard]]},
  {label:"Images",items:[["Gold Images",HardDrive],["ISO Library",Disc3],["Clone Images",Copy],["macOS Installers",Apple]]},
  {label:"Deployment",items:[["Hosts (PXE)",Computer],["Deployments",Rocket],["Task Sequences",ListChecks],["Departments",Building2]]},
  {label:"Management",items:[["Applications",Package],["Drivers",Wrench],["Directory Services",ShieldCheck],["PXE / Network",Network]]},
  {label:"Monitoring",items:[["Audit Logs",ClipboardList],["Reports",BarChart3]]},
  {label:"System",items:[["Admin Center",UserCog],["Settings",Settings]]}
] as const;
async function req(path:string,init?:RequestInit){const r=await fetch(API+path,{...init,credentials:"include",headers:{"Content-Type":"application/json",...(init?.headers||{})}});if(!r.ok){let m=r.status+" "+r.statusText;try{const b=await r.json();m=b.detail||b.message||m}catch{}throw new Error(m)}return r.status===204?null:r.json()}
const img0={name:"",os_name:"Windows 11 Enterprise",os_version:"",architecture:"x86_64",image_version:"1.0",image_path:"",sysprep_ready:false,checksum:"",notes:""};
const sw0={name:"",version:"",installer_path:"",silent_install:"",detection_rule:"",uninstall_command:"",install_order:100,required_by_default:false};
const host0={hostname:"",mac_address:"",serial_number:"",manufacturer:"",model:"",department_id:""};
const dept0={name:"",code:"",image_id:"",computer_name_pattern:"{DEPT}-{SERIAL}",ad_ou:"",required_software_ids:[] as string[],optional_software_ids:[] as string[],printers:"",post_install_scripts:""};
const dir0:Dir={domain:"",domain_controller:"",protocol:"ldaps",port:636,base_dn:"",service_account:"",default_computer_ou:""};

export function App(){
 const [auth,setAuth]=useState<null|boolean>(null),[active,setActive]=useState("Dashboard"),[stats,setStats]=useState<Stats>({images:0,departments:0,software:0,hosts:0,queued_deployments:0}),[images,setImages]=useState<Img[]>([]),[software,setSoftware]=useState<Sw[]>([]),[departments,setDepartments]=useState<Dept[]>([]),[hosts,setHosts]=useState<Host[]>([]),[jobs,setJobs]=useState<Job[]>([]),[directory,setDirectory]=useState<Dir>(dir0),[error,setError]=useState(""),[notice,setNotice]=useState(""),[loading,setLoading]=useState(true);
 const [theme,setTheme]=useState<"light"|"dark"|"system">(()=>((localStorage.getItem("reforge-theme") as any)||"system"));
 async function refresh(silent=false){if(!silent)setLoading(true);try{const a=await Promise.all([req("/api/dashboard"),req("/api/images"),req("/api/software"),req("/api/departments"),req("/api/hosts"),req("/api/deployments"),req("/api/directory")]);setStats(a[0]);setImages(a[1]);setSoftware(a[2]);setDepartments(a[3]);setHosts(a[4]);setJobs(a[5]);setDirectory(a[6]);setError("")}catch(e:any){setError(e.message)}finally{if(!silent)setLoading(false)}}
 useEffect(()=>{req("/api/auth/me").then(()=>{setAuth(true);refresh()}).catch(()=>setAuth(false))},[]);
 useEffect(()=>{if(!auth)return;const t=setInterval(()=>refresh(true),4000);return()=>clearInterval(t)},[auth]);
 useEffect(()=>{localStorage.setItem("reforge-theme",theme);document.documentElement.dataset.theme=theme},[theme]);
 let page:any=<Dashboard stats={stats} jobs={jobs} hosts={hosts} images={images}/>;
 if(active==="Gold Images")page=<Images items={images} changed={refresh} err={setError} note={setNotice}/>;
 if(active==="Applications")page=<Software items={software} changed={refresh} err={setError} note={setNotice}/>;
 if(active==="Departments")page=<Departments items={departments} images={images} software={software} changed={refresh} err={setError} note={setNotice}/>;
 if(active==="Hosts (PXE)")page=<Hosts items={hosts} departments={departments} changed={refresh} err={setError} note={setNotice}/>;
 if(active==="Deployments")page=<Deployments items={jobs} hosts={hosts} images={images} departments={departments} software={software} changed={refresh} err={setError} note={setNotice}/>;
 if(active==="Directory Services")page=<Directory value={directory} changed={refresh} err={setError} note={setNotice}/>;
 if(active==="PXE / Network")page=<PXE/>;
 if(active==="Audit Logs")page=<AuditPage/>;
 if(active==="Settings")page=<SettingsPage/>;
 if(active==="ISO Library")page=<ModulePage icon={Disc3} title="ISO Library" description="Import and manage Windows, Linux and recovery installation media for PXE deployment." action="Import ISO" bullets={["Checksum and architecture metadata","UEFI/BIOS compatibility","Unattended install configuration"]}/>;
 if(active==="Clone Images")page=<ModulePage icon={Copy} title="Clone Images" description="Capture machine-oriented disk images for labs, kiosks and identical hardware fleets." action="Capture clone" bullets={["Disk and partition topology","Hardware-family matching","Restore validation"]}/>;
 if(active==="macOS Installers")page=<ModulePage icon={Apple} title="macOS Provisioning" description="Manage macOS installers, packages, profiles and supported Apple provisioning workflows." action="Add installer" bullets={["Intel and Apple Silicon inventory","PKG and configuration profiles","Apple enrollment integration"]}/>;
 if(active==="Task Sequences")page=<ModulePage icon={ListChecks} title="Task Sequences" description="Build ordered deployment workflows for imaging, drivers, applications, directory join, updates and validation." action="New task sequence" bullets={["Conditional deployment steps","Retries and timeouts","Windows, Linux and macOS filters"]}/>;
 if(active==="Drivers")page=<ModulePage icon={Wrench} title="Driver Packs" description="Organize driver packs by manufacturer, model, operating system and architecture." action="Add driver pack" bullets={["Model matching","Offline Windows injection","Versioned driver packs"]}/>;
 if(active==="Reports")page=<ModulePage icon={BarChart3} title="Reports" description="Deployment, image, host, driver and imaging-node operational reporting." action="Create report" bullets={["Deployment success and duration","Image usage and age","Hardware and OS inventory"]}/>;
 if(active==="Admin Center")page=<AdminCenter go={setActive}/>;
 if(active==="Imaging Nodes")page=<ModulePage icon={Server} title="Imaging Nodes" description="Register and monitor privileged capture and restore workers." action="Register imaging node" bullets={["Signed deployment authorization","Node health and workload state","PXE imaging environment handoff"]}/>;
 if(active==="Storage Nodes")page=<ModulePage icon={Database} title="Storage Nodes" description="Manage repositories for ISO, gold, clone and macOS deployment content." action="Add storage node" bullets={["Local, SMB, NFS and S3-compatible targets","Capacity and checksum health","Replication and retention"]}/>;
 if(auth===null)return <div className="loadingScreen"><Loader2 className="spin"/> Starting ReForge...</div>;
 if(!auth)return <Login onSuccess={()=>{setAuth(true);refresh()}}/>;
 async function logout(){try{await req("/api/auth/logout",{method:"POST"})}finally{setAuth(false)}}
 return <div className="shell">
  <aside>
   <div className="brand"><div className="brandmark"><Boxes size={22}/></div><div><strong>ReForge</strong><span>Deployment Platform</span></div></div>
   <nav>{navGroups.map(group=><div className="navGroup" key={group.label||"main"}>{group.label&&<div className="navLabel">{group.label}</div>}{group.items.map(([l,I])=><button key={l} className={active===l?"active":""} onClick={()=>setActive(l)}><I size={18}/><span>{l}</span></button>)}</div>)}</nav>
   <div className="accountCard"><div className="avatar">A</div><div><b>Administrator</b><span>Local admin</span></div><button onClick={logout} title="Sign out"><LogOut size={16}/></button></div>
  </aside>
  <div className="workspace">
   <div className="topbar">
    <div className="globalSearch"><Search size={17}/><input aria-label="Global search" placeholder="Search images, devices, deployments..."/><kbd>Ctrl K</kbd></div>
    <div className="topActions">
     <div className="themeSwitch"><button className={theme==="light"?"active":""} onClick={()=>setTheme("light")}><Sun size={15}/> Light</button><button className={theme==="dark"?"active":""} onClick={()=>setTheme("dark")}><Moon size={15}/> Dark</button><button className={theme==="system"?"active":""} onClick={()=>setTheme("system")}><Monitor size={15}/> System</button></div>
     <button className="topIcon" title="Notifications"><Bell size={18}/></button><button className="topIcon" title="Help"><HelpCircle size={18}/></button>
    </div>
   </div>
   <main>
    <header className="pageHeader"><div><h1>{active}</h1><p>{active==="Dashboard"?"Manage and deploy systems across your environment.":"ReForge deployment management"}</p></div><button className="secondary" onClick={()=>refresh()}><RefreshCw size={15}/> Refresh</button></header>
    {error&&<Alert kind="error" text={error} close={()=>setError("")}/>} {notice&&<Alert kind="success" text={notice} close={()=>setNotice("")}/>}
    {loading?<div className="loading"><Loader2 className="spin"/> Loading ReForge...</div>:page}
   </main>
  </div>
 </div>
}

function Dashboard(p:{stats:Stats;jobs:Job[];hosts:Host[];images:Img[]}){
 const recent=p.jobs.slice(0,5);
 return <div className="dashboardV2">
  <section className="systemStrip"><div className="serverIdentity"><div className="serverIcon"><Server size={21}/></div><div><b>ReForge Server</b><span>Deployment control plane</span></div></div><div className="systemReady"><span className="statusDot"/><div><b>System Ready</b><span>API and database available</span></div></div></section>
  <section className="panel pipelinePanel"><div className="sectionHead"><div><h3>Deployment Pipeline</h3><p>Step-by-step workflow for deploying systems.</p></div><button className="primary"><Play size={16}/> Start Deployment</button></div><div className="pipelineFlow">{["PXE Boot","Select Source","Drivers","Join Directory","Applications","Complete"].map((x,i)=><div className="pipeStage" key={x}><div className={i<2?"pipeDot done":"pipeDot"}>{i<2?<CheckCircle2 size={18}/>:i+1}</div><b>{x}</b><span>{["Boot device from network","ISO, gold or clone image","Inject matching drivers","AD / Entra placement","Install software and updates","Validate and hand off"][i]}</span></div>)}</div></section>
  <section className="dashboardGrid">
   <article className="panel"><div className="sectionHead"><div><h3>Live Queue</h3><p>Active and recent deployments.</p></div><span className="linkText">{p.stats.queued_deployments} queued · {p.stats.running_deployments||0} running</span></div><table><thead><tr><th>Device</th><th>Image</th><th>Stage</th><th>Progress</th></tr></thead><tbody>{recent.map(j=><tr key={j.id}><td><b>{p.hosts.find(h=>h.id===j.host_id)?.hostname||j.host_id.slice(0,8)}</b></td><td>{p.images.find(i=>i.id===j.image_id)?.name||"Unassigned"}</td><td><Status v={j.status}/></td><td><Progress v={j.progress}/></td></tr>)}{!recent.length&&<Empty cols={4}/>}</tbody></table></article>
   <article className="panel activityPanel"><div className="sectionHead"><div><h3>Environment</h3><p>Managed resources.</p></div></div><div className="metricRows"><div><HardDrive size={18}/><span>Gold Images</span><b>{p.stats.images}</b></div><div><Computer size={18}/><span>Managed Hosts</span><b>{p.stats.hosts}</b></div><div><Building2 size={18}/><span>Departments</span><b>{p.stats.departments}</b></div><div><Package size={18}/><span>Applications</span><b>{p.stats.software}</b></div></div></article>
  </section>
 </div>
}

function Images(p:any){const [f,setF]=useState<any>(img0),[edit,setEdit]=useState<string|null>(null);async function save(e:FormEvent){e.preventDefault();try{await req(edit?"/api/images/"+edit:"/api/images",{method:edit?"PUT":"POST",body:JSON.stringify(f)});setF(img0);setEdit(null);p.note("Gold image saved");p.changed()}catch(e:any){p.err(e.message)}}async function del(id:string){if(confirm("Delete this gold image?"))try{await req("/api/images/"+id,{method:"DELETE"});p.changed()}catch(e:any){p.err(e.message)}}return <Crud title="Gold Images" help="Maintain generalized, versioned Windows or Linux source images." form={<form onSubmit={save} className="formGrid"><F l="Image name"><input required value={f.name} onChange={e=>setF({...f,name:e.target.value})}/></F><F l="Operating system"><input value={f.os_name} onChange={e=>setF({...f,os_name:e.target.value})}/></F><F l="OS version/build"><input value={f.os_version} onChange={e=>setF({...f,os_version:e.target.value})}/></F><F l="Image version"><input value={f.image_version} onChange={e=>setF({...f,image_version:e.target.value})}/></F><F l="Storage path"><input value={f.image_path||""} onChange={e=>setF({...f,image_path:e.target.value})}/></F><label className="check"><input type="checkbox" checked={f.sysprep_ready} onChange={e=>setF({...f,sysprep_ready:e.target.checked})}/> Sysprep / generalized</label><Save editing={!!edit} cancel={()=>{setEdit(null);setF(img0)}}/></form>}><Cards>{p.items.map((x:Img)=><Card key={x.id} title={x.name} sub={x.os_name+" "+x.os_version+" • v"+x.image_version}><div className="meta"><span>{x.architecture}</span><span>{x.sysprep_ready?"Sysprep ready":"Not generalized"}</span></div><p>{x.image_path||"No storage path"}</p><Actions edit={()=>{setEdit(x.id);setF({...x})}} del={()=>del(x.id)}/></Card>)}</Cards></Crud>}

function Software(p:any){const [f,setF]=useState<any>(sw0),[edit,setEdit]=useState<string|null>(null);async function save(e:FormEvent){e.preventDefault();try{await req(edit?"/api/software/"+edit:"/api/software",{method:edit?"PUT":"POST",body:JSON.stringify({...f,install_order:Number(f.install_order)})});setF(sw0);setEdit(null);p.note("Software saved");p.changed()}catch(e:any){p.err(e.message)}}async function seed(){try{const r=await req("/api/software/bootstrap",{method:"POST"});p.note(String(r.created)+" starter packages added");p.changed()}catch(e:any){p.err(e.message)}}async function del(id:string){if(confirm("Delete this package?"))try{await req("/api/software/"+id,{method:"DELETE"});p.changed()}catch(e:any){p.err(e.message)}}return <Crud title="Software Catalog" help="Store installers and silent install logic once, then assign packages to departments." extra={<button className="secondary" onClick={seed}><Database size={15}/> Load starter catalog</button>} form={<form onSubmit={save} className="formGrid"><F l="Application"><input required value={f.name} onChange={e=>setF({...f,name:e.target.value})}/></F><F l="Version"><input value={f.version} onChange={e=>setF({...f,version:e.target.value})}/></F><F l="Installer path"><input value={f.installer_path} onChange={e=>setF({...f,installer_path:e.target.value})}/></F><F l="Install order"><input type="number" value={f.install_order} onChange={e=>setF({...f,install_order:e.target.value})}/></F><F l="Silent command" wide><input value={f.silent_install} onChange={e=>setF({...f,silent_install:e.target.value})}/></F><F l="Detection rule" wide><input value={f.detection_rule} onChange={e=>setF({...f,detection_rule:e.target.value})}/></F><Save editing={!!edit} cancel={()=>{setEdit(null);setF(sw0)}}/></form>}><table><thead><tr><th>Application</th><th>Version</th><th>Order</th><th>Silent command</th><th></th></tr></thead><tbody>{p.items.map((x:Sw)=><tr key={x.id}><td><b>{x.name}</b></td><td>{x.version||"—"}</td><td>{x.install_order}</td><td className="mono">{x.silent_install||"Not configured"}</td><td><Actions edit={()=>{setEdit(x.id);setF({...x})}} del={()=>del(x.id)}/></td></tr>)}{!p.items.length&&<Empty cols={5}/>}</tbody></table></Crud>}

function Departments(p:any){const [f,setF]=useState<any>(dept0),[edit,setEdit]=useState<string|null>(null);function tog(k:string,id:string){const a=f[k] as string[];setF({...f,[k]:a.includes(id)?a.filter(x=>x!==id):a.concat(id)})}async function save(e:FormEvent){e.preventDefault();try{const body={...f,image_id:f.image_id||null,printers:String(f.printers).split("\n").filter(Boolean),post_install_scripts:String(f.post_install_scripts).split("\n").filter(Boolean),config:{}};await req(edit?"/api/departments/"+edit:"/api/departments",{method:edit?"PUT":"POST",body:JSON.stringify(body)});setF(dept0);setEdit(null);p.note("Department profile saved");p.changed()}catch(e:any){p.err(e.message)}}async function del(id:string){if(confirm("Delete department?"))try{await req("/api/departments/"+id,{method:"DELETE"});p.changed()}catch(e:any){p.err(e.message)}}return <Crud title="Department Profiles" help="Layer department configuration and software on top of a reusable gold image." form={<form onSubmit={save} className="formGrid"><F l="Department"><input required value={f.name} onChange={e=>setF({...f,name:e.target.value})}/></F><F l="Code"><input required value={f.code} onChange={e=>setF({...f,code:e.target.value.toUpperCase()})}/></F><F l="Gold image"><select value={f.image_id||""} onChange={e=>setF({...f,image_id:e.target.value})}><option value="">Select image</option>{p.images.map((x:Img)=><option value={x.id} key={x.id}>{x.name}</option>)}</select></F><F l="Computer naming"><input value={f.computer_name_pattern} onChange={e=>setF({...f,computer_name_pattern:e.target.value})}/></F><F l="Active Directory OU" wide><input value={f.ad_ou} onChange={e=>setF({...f,ad_ou:e.target.value})}/></F><F l="Required software" wide><Choices data={p.software} selected={f.required_software_ids} change={(id:string)=>tog("required_software_ids",id)}/></F><F l="Optional software" wide><Choices data={p.software} selected={f.optional_software_ids} change={(id:string)=>tog("optional_software_ids",id)}/></F><F l="Printers"><textarea value={f.printers} onChange={e=>setF({...f,printers:e.target.value})}/></F><F l="Post-install scripts"><textarea value={f.post_install_scripts} onChange={e=>setF({...f,post_install_scripts:e.target.value})}/></F><Save editing={!!edit} cancel={()=>{setEdit(null);setF(dept0)}}/></form>}><Cards>{p.items.map((x:Dept)=><Card key={x.id} title={x.name} sub={x.code+" • "+(p.images.find((i:Img)=>i.id===x.image_id)?.name||"No image")}><p className="mono">{x.ad_ou||"No AD OU"}</p><div className="meta"><span>{x.required_software_ids.length} required apps</span><span>{x.optional_software_ids.length} optional</span></div><Actions edit={()=>{setEdit(x.id);setF({...x,printers:(x.printers||[]).join("\n"),post_install_scripts:(x.post_install_scripts||[]).join("\n")})}} del={()=>del(x.id)}/></Card>)}</Cards></Crud>}

function Hosts(p:any){const [f,setF]=useState<any>(host0),[edit,setEdit]=useState<string|null>(null);async function save(e:FormEvent){e.preventDefault();try{await req(edit?"/api/hosts/"+edit:"/api/hosts",{method:edit?"PUT":"POST",body:JSON.stringify({...f,department_id:f.department_id||null})});setF(host0);setEdit(null);p.note("Host saved");p.changed()}catch(e:any){p.err(e.message)}}async function del(id:string){if(confirm("Delete host?"))try{await req("/api/hosts/"+id,{method:"DELETE"});p.changed()}catch(e:any){p.err(e.message)}}return <Crud title="Managed Hosts" help="Register endpoints manually or automatically through the PXE inventory flow." form={<form onSubmit={save} className="formGrid"><F l="Hostname"><input required value={f.hostname} onChange={e=>setF({...f,hostname:e.target.value})}/></F><F l="MAC address"><input required value={f.mac_address} onChange={e=>setF({...f,mac_address:e.target.value})}/></F><F l="Serial number"><input value={f.serial_number} onChange={e=>setF({...f,serial_number:e.target.value})}/></F><F l="Manufacturer"><input value={f.manufacturer} onChange={e=>setF({...f,manufacturer:e.target.value})}/></F><F l="Model"><input value={f.model} onChange={e=>setF({...f,model:e.target.value})}/></F><F l="Department"><select value={f.department_id||""} onChange={e=>setF({...f,department_id:e.target.value})}><option value="">Unassigned</option>{p.departments.map((d:Dept)=><option key={d.id} value={d.id}>{d.name}</option>)}</select></F><Save editing={!!edit} cancel={()=>{setEdit(null);setF(host0)}}/></form>}><table><thead><tr><th>Hostname</th><th>MAC</th><th>Hardware</th><th>Department</th><th></th></tr></thead><tbody>{p.items.map((x:Host)=><tr key={x.id}><td><b>{x.hostname}</b><small className="block">{x.serial_number}</small></td><td className="mono">{x.mac_address}</td><td>{(x.manufacturer+" "+x.model).trim()||"—"}</td><td>{p.departments.find((d:Dept)=>d.id===x.department_id)?.name||"Unassigned"}</td><td><Actions edit={()=>{setEdit(x.id);setF({...x,department_id:x.department_id||""})}} del={()=>del(x.id)}/></td></tr>)}{!p.items.length&&<Empty cols={5}/>}</tbody></table></Crud>}

function Deployments(p:any){
 const [step,setStep]=useState(0);
 const [f,setF]=useState<any>({host_id:"",image_id:"",department_id:"",optional_software:[]});
 const dept=p.departments.find((d:Dept)=>d.id===f.department_id);
 const host=p.hosts.find((h:Host)=>h.id===f.host_id);
 const image=p.images.find((i:Img)=>i.id===f.image_id);
 const offered=p.software.filter((s:Sw)=>dept?.optional_software_ids.includes(s.id));
 useEffect(()=>{if(dept?.image_id)setF((x:any)=>({...x,image_id:dept.image_id,optional_software:[]}))},[f.department_id]);
 function tog(id:string){setF({...f,optional_software:f.optional_software.includes(id)?f.optional_software.filter((x:string)=>x!==id):f.optional_software.concat(id)})}
 async function queue(){try{await req("/api/deployments",{method:"POST",body:JSON.stringify({...f,department_id:f.department_id||null})});setF({host_id:"",image_id:"",department_id:"",optional_software:[]});setStep(0);p.note("Deployment queued");p.changed()}catch(e:any){p.err(e.message)}}
 async function act(id:string,a:string){try{await req("/api/deployments/"+id+"/"+a,{method:"POST"});p.changed()}catch(e:any){p.err(e.message)}}
 const stages=["Computer","Task sequence","Applications","Review"];
 return <div className="stack">
   <section className="panel deployWizard">
     <div className="wizardRail">{stages.map((s,i)=><button key={s} className={step===i?"wizardStep active":step>i?"wizardStep done":"wizardStep"} onClick={()=>{if(i<step)setStep(i)}}><span>{i+1}</span><div><b>{s}</b><small>{i===0?"Select target device":i===1?"Image and department":i===2?"Optional software":"Confirm and deploy"}</small></div></button>)}</div>
     <div className="wizardBody">
       <div className="wizardTitle"><span>Deployment task sequence</span><h2>{stages[step]}</h2></div>
       {step===0&&<div className="wizardContent"><F l="Target computer"><select value={f.host_id} onChange={e=>setF({...f,host_id:e.target.value})}><option value="">Choose a managed device</option>{p.hosts.map((h:Host)=><option value={h.id} key={h.id}>{h.hostname+" — "+h.mac_address}</option>)}</select></F>{host&&<div className="selectionSummary"><b>{host.hostname}</b><span>{host.manufacturer+" "+host.model}</span><span className="mono">{host.mac_address}</span></div>}</div>}
       {step===1&&<div className="wizardContent formGrid"><F l="Department"><select value={f.department_id} onChange={e=>setF({...f,department_id:e.target.value})}><option value="">No department profile</option>{p.departments.map((d:Dept)=><option value={d.id} key={d.id}>{d.name}</option>)}</select></F><F l="Gold image"><select value={f.image_id} onChange={e=>setF({...f,image_id:e.target.value})}><option value="">Choose image</option>{p.images.map((i:Img)=><option value={i.id} key={i.id}>{i.name+" v"+i.image_version}</option>)}</select></F>{dept&&<div className="wide selectionSummary"><b>{dept.name}</b><span>{"Naming: "+dept.computer_name_pattern}</span><span className="mono">{dept.ad_ou||"No directory OU assigned"}</span></div>}</div>}
       {step===2&&<div className="wizardContent"><div className="taskSequenceList"><div className="taskRow locked"><span>1</span><div><b>Apply gold image</b><small>{image?.name||"Choose an image first"}</small></div></div><div className="taskRow locked"><span>2</span><div><b>Install required department software</b><small>{String(dept?.required_software_ids.length||0)+" packages"}</small></div></div><div className="taskRow"><span>3</span><div><b>Optional applications</b><small>Select what this computer also needs</small></div></div></div><Choices data={offered} selected={f.optional_software} change={tog}/></div>}
       {step===3&&<div className="wizardContent"><div className="reviewGrid"><div><span>Computer</span><b>{host?.hostname||"Not selected"}</b></div><div><span>Gold image</span><b>{image?.name||"Not selected"}</b></div><div><span>Department</span><b>{dept?.name||"None"}</b></div><div><span>Optional apps</span><b>{f.optional_software.length}</b></div><div><span>PXE action</span><b>Boot and deploy</b></div><div><span>Directory</span><b>{dept?.ad_ou?"Join configured OU":"No department OU"}</b></div></div><div className="deployWarning"><ShieldCheck size={18}/><div><b>Ready to queue</b><span>The endpoint will receive this task sequence the next time it PXE boots.</span></div></div></div>}
       <div className="wizardActions"><button className="secondary" disabled={step===0} onClick={()=>setStep(Math.max(0,step-1))}>Back</button>{step<3?<button className="primary" disabled={(step===0&&!f.host_id)||(step===1&&!f.image_id)} onClick={()=>setStep(step+1)}>Next</button>:<button className="primary" disabled={!f.host_id||!f.image_id} onClick={queue}><Play size={15}/> Queue deployment</button>}</div>
     </div>
   </section>
   <section className="panel"><div className="sectionHead"><div><h3>Deployment activity</h3><p>Queued, waiting, running and completed task sequences.</p></div></div><table><thead><tr><th>Host</th><th>Image / Department</th><th>Status</th><th>Progress</th><th>Current step</th><th></th></tr></thead><tbody>{p.items.map((j:Job)=><tr key={j.id}><td><b>{p.hosts.find((h:Host)=>h.id===j.host_id)?.hostname||"Unknown"}</b></td><td>{p.images.find((i:Img)=>i.id===j.image_id)?.name||"Unknown"}<small className="block">{p.departments.find((d:Dept)=>d.id===j.department_id)?.name||"No department"}</small></td><td><Status v={j.status}/></td><td><Progress v={j.progress}/></td><td>{j.current_step}</td><td>{["failed","canceled","waiting"].includes(j.status)&&<button className="iconBtn" onClick={()=>act(j.id,"retry")}><RefreshCw size={15}/></button>}{["queued","running"].includes(j.status)&&<button className="iconBtn danger" onClick={()=>act(j.id,"cancel")}><XCircle size={15}/></button>}</td></tr>)}{!p.items.length&&<Empty cols={6}/>}</tbody></table></section>
 </div>
}
function Directory(p:any){const [f,setF]=useState<Dir>(p.value),[password,setPassword]=useState("");useEffect(()=>setF(p.value),[p.value]);async function save(e:FormEvent){e.preventDefault();try{await req("/api/directory",{method:"PUT",body:JSON.stringify(f)});p.note("Directory settings saved");p.changed()}catch(e:any){p.err(e.message)}}async function test(){try{const r=await req("/api/directory/test",{method:"POST",body:JSON.stringify({...f,password})});p.note(r.message)}catch(e:any){p.err(e.message)}}return <section className="panel"><h3>Active Directory / LDAP</h3><p>Use LDAPS where available. The test password is not stored.</p><form onSubmit={save} className="formGrid formWrap"><F l="Domain"><input value={f.domain} onChange={e=>setF({...f,domain:e.target.value})}/></F><F l="Domain controller"><input value={f.domain_controller} onChange={e=>setF({...f,domain_controller:e.target.value})}/></F><F l="Protocol"><select value={f.protocol} onChange={e=>setF({...f,protocol:e.target.value,port:e.target.value==="ldaps"?636:389})}><option value="ldaps">LDAPS</option><option value="starttls">LDAP + StartTLS</option><option value="ldap">Plain LDAP (disabled by default)</option></select></F><F l="Port"><input type="number" value={f.port} onChange={e=>setF({...f,port:Number(e.target.value)})}/></F><F l="Base DN" wide><input value={f.base_dn} onChange={e=>setF({...f,base_dn:e.target.value})}/></F><F l="Service account"><input value={f.service_account} onChange={e=>setF({...f,service_account:e.target.value})}/></F><F l="Default computer OU"><input value={f.default_computer_ou} onChange={e=>setF({...f,default_computer_ou:e.target.value})}/></F><F l="Password for test"><input type="password" value={password} onChange={e=>setPassword(e.target.value)}/></F><div className="formButtons wide"><button className="primary">Save settings</button><button className="secondary" type="button" disabled={!password} onClick={test}><ShieldCheck size={15}/> Test connection</button></div></form></section>}

function PXE(){
 const [f,setF]=useState<any>(null),[saved,setSaved]=useState(false),[error,setError]=useState("");
 useEffect(()=>{req("/api/pxe").then((x:any)=>setF({
   menu_title:"ReForge Deployment",default_item:"deploy",
   show_deploy:true,show_register:true,show_diagnostics:true,show_local_boot:true,...x
 })).catch((e:any)=>setError(e.message))},[]);
 if(!f)return <section className="panel"><div className="loading"><Loader2 className="spin"/> Loading PXE configuration...</div></section>;
 async function save(){try{const updated=await req("/api/pxe",{method:"PUT",body:JSON.stringify({...f,boot_menu_timeout:Number(f.boot_menu_timeout)})});setF(updated);setSaved(true);setTimeout(()=>setSaved(false),2500)}catch(e:any){setError(e.message)}}
 const bootURL=(f.server_url||window.location.origin)+"/boot/ipxe";
 const entries=[
   ["deploy","Deploy assigned image",Boolean(f.show_deploy)],
   ["register","Register this device",Boolean(f.show_register)],
   ["diagnostics","Diagnostics and inventory",Boolean(f.show_diagnostics)],
   ["local","Boot local disk",Boolean(f.show_local_boot)]
 ] as const;
 const visible=entries.filter(x=>x[2]);
 function toggle(k:string,v:boolean){
   const next={...f,[k]:v};
   const nextEntries=[
    ["deploy","show_deploy"],["register","show_register"],["diagnostics","show_diagnostics"],["local","show_local_boot"]
   ].filter(([,key])=>Boolean(next[key]));
   if(!nextEntries.some(([id])=>id===next.default_item)) next.default_item=nextEntries[0]?.[0]||"local";
   setF(next);
 }
 return <div className="stack">
  <section className="panel"><div className="sectionHead"><div><h3>PXE boot service</h3><p>Configure ReForge to work with your existing DHCP service or a dedicated imaging network.</p></div><label className="switch"><input type="checkbox" checked={f.enabled} onChange={e=>setF({...f,enabled:e.target.checked})}/><span>{f.enabled?"Enabled":"Disabled"}</span></label></div>{error&&<div className="alert error">{error}</div>}<div className="formGrid formWrap"><F l="ReForge server URL"><input value={f.server_url} onChange={e=>setF({...f,server_url:e.target.value})}/></F><F l="DHCP integration"><select value={f.dhcp_mode} onChange={e=>setF({...f,dhcp_mode:e.target.value})}><option value="existing">Use existing DHCP server</option><option value="proxy">Proxy DHCP / imaging VLAN</option><option value="managed">ReForge-managed DHCP</option></select></F><F l="DHCP server"><input placeholder="10.0.0.10" value={f.dhcp_server||""} onChange={e=>setF({...f,dhcp_server:e.target.value})}/></F><F l="Next server / TFTP"><input placeholder="10.0.0.20" value={f.next_server||""} onChange={e=>setF({...f,next_server:e.target.value,tftp_server:e.target.value})}/></F><F l="Legacy BIOS boot file"><input value={f.bios_boot_file} onChange={e=>setF({...f,bios_boot_file:e.target.value})}/></F><F l="UEFI x64 boot file"><input value={f.uefi_boot_file} onChange={e=>setF({...f,uefi_boot_file:e.target.value})}/></F><label className="check"><input type="checkbox" checked={f.allow_unknown} onChange={e=>setF({...f,allow_unknown:e.target.checked})}/> Allow unknown devices to reach registration</label><div></div></div></section>

  <section className="pxeDesignerGrid">
   <article className="panel">
    <div className="sectionHead"><div><h3>Boot Menu Designer</h3><p>Customize what users see when a device PXE boots.</p></div></div>
    <div className="formGrid formWrap">
     <F l="Menu title" wide><input value={f.menu_title||""} onChange={e=>setF({...f,menu_title:e.target.value})}/></F>
     <F l="Default selection"><select value={f.default_item||"deploy"} onChange={e=>setF({...f,default_item:e.target.value})}>{visible.map(([id,label])=><option key={id} value={id}>{label}</option>)}</select></F>
     <F l="Timeout (seconds)"><input type="number" min="1" max="60" value={f.boot_menu_timeout} onChange={e=>setF({...f,boot_menu_timeout:Number(e.target.value)})}/></F>
     <div className="wide bootOptionGrid">
      <label><input type="checkbox" checked={Boolean(f.show_deploy)} onChange={e=>toggle("show_deploy",e.target.checked)}/><span><b>Deploy</b><small>Execute assigned image/task</small></span></label>
      <label><input type="checkbox" checked={Boolean(f.show_register)} onChange={e=>toggle("show_register",e.target.checked)}/><span><b>Register</b><small>Enroll an unknown device</small></span></label>
      <label><input type="checkbox" checked={Boolean(f.show_diagnostics)} onChange={e=>toggle("show_diagnostics",e.target.checked)}/><span><b>Diagnostics</b><small>Inventory and troubleshooting</small></span></label>
      <label><input type="checkbox" checked={Boolean(f.show_local_boot)} onChange={e=>toggle("show_local_boot",e.target.checked)}/><span><b>Local Boot</b><small>Continue to installed OS</small></span></label>
     </div>
     <div className="wide formButtons"><button className="primary" onClick={save}>Save PXE settings</button>{saved&&<span className="savedText"><CheckCircle2 size={15}/> Saved and active</span>}</div>
    </div>
   </article>

   <article className="panel bootPreviewPanel">
    <div className="sectionHead"><div><h3>Live boot preview</h3><p>Preview updates before you save.</p></div><span className="previewBadge">iPXE</span></div>
    <div className="bootScreen">
     <div className="bootLogo">ReForge</div>
     <div className="bootTitle">{f.menu_title||"ReForge Deployment"}</div>
     <div className="bootMenu">{visible.length?visible.map(([id,label],i)=><div className={"bootMenuItem "+((f.default_item||"deploy")===id?"selected":"")} key={id}><span>{((f.default_item||"deploy")===id)?"›":" "}</span><b>{label}</b><small>{id==="deploy"?"Deploy the assigned ISO, gold or clone image":id==="register"?"Register this device with ReForge":id==="diagnostics"?"Run hardware inventory and diagnostics":"Boot from the local disk"}</small></div>):<div className="bootEmpty">At least one boot option is required.</div>}</div>
     <div className="bootFooter">Automatic selection in {Number(f.boot_menu_timeout)||5}s · ↑ ↓ select · Enter continue</div>
    </div>
   </article>
  </section>

  <section className="panel"><h3>Boot configuration</h3><p>Point DHCP/PXE to the boot loader and chain into ReForge.</p><div className="pxeSummary"><div><span>BIOS</span><b>{f.bios_boot_file}</b></div><div><span>UEFI x64</span><b>{f.uefi_boot_file}</b></div><div><span>Next server</span><b>{f.next_server||"Set your ReForge/TFTP server"}</b></div><div><span>ReForge chain URL</span><b className="mono">{bootURL}</b></div></div><div className="codeBox">curl {bootURL}</div></section>
  <section className="panel"><h3>PXE validation checklist</h3><div className="checkList">{["DHCP reachable from deployment VLAN","BIOS and UEFI boot files available","HTTP access to ReForge web/API","Unknown-device registration policy reviewed","Imaging node attached before disk operations"].map((x,i)=><div key={x}><span>{i+1}</span><b>{x}</b></div>)}</div></section>
 </div>
}
function SettingsPage(){
 const [health,setHealth]=useState<any>(null),[current,setCurrent]=useState(""),[next,setNext]=useState(""),[confirm,setConfirm]=useState(""),[msg,setMsg]=useState("");
 useEffect(()=>{fetch("/health").then(r=>r.json()).then(setHealth).catch(()=>{})},[]);
 async function changePassword(e:FormEvent){e.preventDefault();setMsg("");if(next!==confirm){setMsg("New passwords do not match");return}try{await req("/api/auth/password",{method:"POST",body:JSON.stringify({current_password:current,new_password:next})});setCurrent("");setNext("");setConfirm("");setMsg("Password changed")}catch(e:any){setMsg(e.message)}}
 return <div className="stack">
  <section className="panel"><h3>Server</h3><div className="settingsRows"><div><span>API status</span><b>{health?.status||"Checking..."}</b></div><div><span>Version</span><b>{health?.version||"—"}</b></div><div><span>Primary database</span><b>{health?.database||"—"}</b></div><div><span>Runtime</span><b>Go control plane + Go worker</b></div></div></section>
  <section className="panel"><h3>Administrator password</h3><p>Rotate the local administrator password. Use SSO/MFA when identity federation is added.</p><form className="formGrid formWrap" onSubmit={changePassword}><F l="Current password"><input type="password" value={current} onChange={e=>setCurrent(e.target.value)}/></F><div></div><F l="New password"><input type="password" minLength={12} value={next} onChange={e=>setNext(e.target.value)}/></F><F l="Confirm new password"><input type="password" minLength={12} value={confirm} onChange={e=>setConfirm(e.target.value)}/></F><div className="formButtons wide"><button className="primary"><KeyRound size={15}/> Change password</button>{msg&&<span className="savedText">{msg}</span>}</div></form></section>
  <section className="panel"><h3>Security posture</h3><div className="tagCloud">{["Authenticated admin API","Argon2id passwords","Audit events","Secure cookies","Exact-origin CORS","LDAPS / StartTLS","Non-root containers","Dependency scans","Secret scans","Isolated imaging node"].map(x=><span key={x}>{x}</span>)}</div></section>
 </div>
}

function AuditPage(){
 const [rows,setRows]=useState<any[]>([]),[error,setError]=useState("");
 useEffect(()=>{req("/api/audit").then(setRows).catch((e:any)=>setError(e.message))},[]);
 return <section className="panel"><div className="sectionHead"><div><h3>Audit log</h3><p>Administrative security and configuration activity recorded by ReForge.</p></div></div>{error&&<div className="alert error">{error}</div>}<table><thead><tr><th>Time</th><th>Actor</th><th>Action</th><th>Resource</th><th>Source</th><th>Result</th></tr></thead><tbody>{rows.map(x=><tr key={x.id}><td>{new Date(x.at).toLocaleString()}</td><td><b>{x.actor}</b></td><td>{x.action}</td><td>{x.resource}{x.resource_id?" / "+x.resource_id.slice(0,8):""}</td><td className="mono">{x.remote_ip}</td><td><span className={"status "+(x.success?"succeeded":"failed")}>{x.success?"Success":"Failed"}</span></td></tr>)}{!rows.length&&<Empty cols={6}/>}</tbody></table></section>
}

function ModulePage(p:{icon:any;title:string;description:string;action:string;bullets:string[]}){
 const I=p.icon;
 return <div className="stack"><section className="panel moduleIntro"><div className="moduleTitle"><div className="moduleIcon"><I size={22}/></div><div><h3>{p.title}</h3><p>{p.description}</p></div></div><button className="primary" disabled title="Backend API is the next implementation phase"><Plus size={15}/>{p.action}</button></section><section className="panel emptyModule"><div className="emptyModuleIcon"><I size={28}/></div><h3>No records yet</h3><p>This v2 management surface is now in the server UI. The next implementation phase will add its database model, API and imaging-node workflow.</p><div className="featureChecks">{p.bullets.map(x=><span key={x}><CheckCircle2 size={15}/>{x}</span>)}</div></section></div>
}

function AdminCenter(p:{go:(page:string)=>void}){
 const tools=[
  ["Users & Roles",UserCog,"Accounts, password and session administration","Settings"],
  ["Authentication",KeyRound,"Local login, AD/LDAP and directory authentication","Directory Services"],
  ["Imaging Nodes",Server,"Privileged capture and restore workers","Imaging Nodes"],
  ["PXE Infrastructure",Network,"Boot services, DHCP integration and boot menu designer","PXE / Network"],
  ["Storage",Database,"Image repositories, capacity and replication","Storage Nodes"],
  ["Security & Audit",ShieldCheck,"Security controls and administrative activity","Audit Logs"]
 ];
 return <div className="adminGrid">{tools.map(([title,I,desc,target]:any)=><article className="panel adminTile" key={title}><I size={21}/><h3>{title}</h3><p>{desc}</p><button className="secondary" onClick={()=>p.go(target)}>Configure</button></article>)}</div>
}

function Crud(p:any){const [show,setShow]=useState(false);return <div className="stack"><section className="panel"><div className="sectionHead"><div><h3>{p.title}</h3><p>{p.help}</p></div><div className="toolbar">{p.extra}<button className="primary" onClick={()=>setShow(!show)}><Plus size={15}/>{show?"Close":"Add new"}</button></div></div>{show&&<div className="formWrap">{p.form}</div>}</section><section className="panel">{p.children}</section></div>}
function F(p:{l:string;wide?:boolean;children:any}){return <label className={p.wide?"field wide":"field"}><span>{p.l}</span>{p.children}</label>}
function Save(p:{editing:boolean;cancel:()=>void}){return <div className="formButtons wide"><button className="primary" type="submit">{p.editing?"Save changes":"Create"}</button>{p.editing&&<button className="secondary" type="button" onClick={p.cancel}>Cancel edit</button>}</div>}
function Choices(p:{data:Sw[];selected:string[];change:(id:string)=>void}){return <div className="choiceGrid">{p.data.length?p.data.map(s=><label className="choice" key={s.id}><input type="checkbox" checked={p.selected.includes(s.id)} onChange={()=>p.change(s.id)}/><span><b>{s.name}</b><small>{s.version||"Package"}</small></span></label>):<span className="muted">No software packages yet.</span>}</div>}
function Cards(p:any){return <div className="cards">{p.children}</div>}
function Card(p:any){return <article className="itemCard"><div><h4>{p.title}</h4><small>{p.sub}</small></div>{p.children}</article>}
function Actions(p:{edit:()=>void;del:()=>void}){return <div className="rowActions"><button onClick={p.edit}>Edit</button><button className="danger" onClick={p.del}><Trash2 size={14}/> Delete</button></div>}
function Empty(p:{cols:number}){return <tr><td colSpan={p.cols} className="emptyCell">No records yet.</td></tr>}
function Status(p:{v:string}){return <span className={"status "+p.v}>{p.v}</span>}
function Progress(p:{v:number}){return <div className="progress"><div style={{width:String(p.v)+"%"}}/><span>{p.v}%</span></div>}
function Alert(p:{kind:string;text:string;close:()=>void}){return <div className={"alert "+p.kind}>{p.kind==="success"?<CheckCircle2 size={17}/>:<XCircle size={17}/>}<span>{p.text}</span><button onClick={p.close}>×</button></div>}


function Login(p:{onSuccess:()=>void}){
 const [username,setUsername]=useState("admin"),[password,setPassword]=useState(""),[error,setError]=useState(""),[busy,setBusy]=useState(false);
 async function submit(e:FormEvent){e.preventDefault();setBusy(true);setError("");try{const r=await fetch(API+"/api/auth/login",{method:"POST",credentials:"include",headers:{"Content-Type":"application/json"},body:JSON.stringify({username,password})});if(!r.ok){const b=await r.json().catch(()=>({}));throw new Error(b.detail||"Login failed")}p.onSuccess()}catch(e:any){setError(e.message)}finally{setBusy(false)}}
 return <div className="loginPage"><div className="loginShell"><div className="loginBrandLarge"><div className="brandmark"><Boxes size={24}/></div><div><strong>ReForge</strong><span>Deployment Platform</span></div></div><form className="loginCard" onSubmit={submit}><h1>Sign in</h1><p>Use your administrator account to continue.</p>{error&&<div className="loginError">{error}</div>}<label><span>Username</span><input autoFocus autoComplete="username" value={username} onChange={e=>setUsername(e.target.value)}/></label><label><span>Password</span><input type="password" autoComplete="current-password" value={password} onChange={e=>setPassword(e.target.value)}/></label><div className="loginOptions"><label className="remember"><input type="checkbox"/> Remember me</label><button type="button">Forgot password?</button></div><button className="primary loginButton" disabled={busy||!username||!password}>{busy?<Loader2 className="spin" size={16}/>:null} Sign in</button></form><span className="loginFoot">ReForge · Secure endpoint deployment</span></div></div>
}
