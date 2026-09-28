import { useEffect, useState } from "react";
import {
  Boxes, Building2, Computer, HardDrive, LayoutDashboard,
  Package, Rocket, Settings, ShieldCheck
} from "lucide-react";

const API = import.meta.env.VITE_API_URL || "http://localhost:8080";

type Stats = {
  images:number; departments:number; software:number; hosts:number; queued_deployments:number;
};

const nav = [
  ["Dashboard", LayoutDashboard],
  ["Gold Images", HardDrive],
  ["Departments", Building2],
  ["Software", Package],
  ["Hosts", Computer],
  ["Deployments", Rocket],
  ["Directory Services", ShieldCheck],
  ["Settings", Settings],
] as const;

export function App() {
  const [active, setActive] = useState("Dashboard");
  const [stats, setStats] = useState<Stats>({
    images:0, departments:0, software:0, hosts:0, queued_deployments:0
  });

  useEffect(() => {
    fetch(`${API}/api/dashboard`).then(r => r.json()).then(setStats).catch(() => {});
  }, []);

  return <div className="shell">
    <aside>
      <div className="brand"><div className="brandmark"><Boxes size={21}/></div>
        <div><strong>ReForge</strong><span>Deployment Platform</span></div>
      </div>
      <nav>{nav.map(([label, Icon]) =>
        <button key={label} className={active===label?"active":""} onClick={()=>setActive(label)}>
          <Icon size={18}/><span>{label}</span>
        </button>
      )}</nav>
      <div className="version">ReForge v0.1.0</div>
    </aside>

    <main>
      <header>
        <div><span className="eyebrow">CONTROL PLANE</span><h1>{active}</h1></div>
        <button className="primary">+ New deployment</button>
      </header>

      {active === "Dashboard" ? <Dashboard stats={stats}/> :
        <section className="panel empty">
          <div className="iconCircle">{nav.find(n=>n[0]===active)?.[1]({size:26} as any)}</div>
          <h2>{active}</h2>
          <p>This module is wired into the ReForge navigation and API foundation. Its full management workflow is next.</p>
        </section>}
    </main>
  </div>
}

function Dashboard({stats}:{stats:Stats}) {
  const cards = [
    ["Gold Images", stats.images, "Ready for deployment"],
    ["Departments", stats.departments, "Deployment profiles"],
    ["Software Packages", stats.software, "Catalog items"],
    ["Managed Hosts", stats.hosts, "Registered endpoints"],
  ];
  return <>
    <section className="hero">
      <div><span className="statusDot"/>SYSTEM READY</div>
      <h2>Build once. Deploy everywhere.</h2>
      <p>Gold images, department profiles, application packages, and directory integration in one deployment control plane.</p>
    </section>

    <section className="stats">
      {cards.map(([name,value,sub])=><article key={name as string}>
        <span>{name}</span><strong>{value}</strong><small>{sub}</small>
      </article>)}
    </section>

    <section className="grid">
      <article className="panel">
        <div className="panelTitle"><div><h3>Deployment Pipeline</h3><p>Standard endpoint provisioning flow</p></div></div>
        <div className="pipeline">
          {["PXE Boot","Gold Image","Drivers","Department","Applications","Join AD","Updates","Complete"].map((x,i)=>
            <div className="step" key={x}><b>{String(i+1).padStart(2,"0")}</b><span>{x}</span></div>
          )}
        </div>
      </article>
      <article className="panel queue">
        <h3>Deployment Queue</h3>
        <div className="queueNumber">{stats.queued_deployments}</div>
        <p>Jobs waiting to run</p>
        <button>View deployments</button>
      </article>
    </section>
  </>;
}
