import React from 'react';
import { NavLink } from 'react-router-dom';
import Icon from '../icons';
import { navSections } from '../../../routes/navItems';
import './Sidebar.css';

interface SidebarProps {
  collapsed: boolean;
  mobileOpen: boolean;
  onNavigate: () => void;
  onBackdropClick: () => void;
}

const Sidebar: React.FC<SidebarProps> = ({
  collapsed,
  mobileOpen,
  onNavigate,
  onBackdropClick,
}) => (
  <>
    {mobileOpen && (
      <div className="sidebar__backdrop" onClick={onBackdropClick} />
    )}
    <nav
      className={[
        'sidebar',
        collapsed ? 'sidebar--collapsed' : '',
        mobileOpen ? 'sidebar--mobile-open' : '',
      ]
        .filter(Boolean)
        .join(' ')}
      aria-label="Primary"
    >
      {navSections.map((section, index) => (
        <div className="sidebar__section" key={section.title ?? index}>
          {section.title && (
            <div className="sidebar__section-title">{section.title}</div>
          )}
          {section.items.map((item) => (
            <NavLink
              key={item.path}
              to={item.path}
              end
              className={({ isActive }) =>
                `sidebar__link${isActive ? ' sidebar__link--active' : ''}`
              }
              onClick={onNavigate}
              title={collapsed ? item.label : undefined}
            >
              <span className="sidebar__icon">
                <Icon name={item.icon} />
              </span>
              <span className="sidebar__label">{item.label}</span>
            </NavLink>
          ))}
        </div>
      ))}
    </nav>
  </>
);

export default Sidebar;
