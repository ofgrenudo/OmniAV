import React, { useState } from 'react';
import { Outlet } from 'react-router-dom';
import Topbar from '../Topbar';
import Sidebar from '../Sidebar';
import './AppLayout.css';

const AppLayout: React.FC = () => {
  const [collapsed, setCollapsed] = useState(false);
  const [mobileOpen, setMobileOpen] = useState(false);

  return (
    <div
      className={`app-layout${collapsed ? ' app-layout--collapsed' : ''}`}
    >
      <Topbar
        collapsed={collapsed}
        onMenuClick={() => setMobileOpen((open) => !open)}
        onCollapseClick={() => setCollapsed((c) => !c)}
      />
      <Sidebar
        collapsed={collapsed}
        mobileOpen={mobileOpen}
        onNavigate={() => setMobileOpen(false)}
        onBackdropClick={() => setMobileOpen(false)}
      />
      <main className="app-layout__content">
        <Outlet />
      </main>
    </div>
  );
};

export default AppLayout;
