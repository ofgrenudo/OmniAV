import React from 'react';
import Icon from '../icons';
import './Topbar.css';

interface TopbarProps {
  collapsed: boolean;
  onMenuClick: () => void;
  onCollapseClick: () => void;
}

const Topbar: React.FC<TopbarProps> = ({
  collapsed,
  onMenuClick,
  onCollapseClick,
}) => (
  <header className="topbar">
    <button
      type="button"
      className="topbar__icon-btn topbar__menu-btn"
      onClick={onMenuClick}
      aria-label="Toggle navigation menu"
    >
      <Icon name="menu" />
    </button>
    <button
      type="button"
      className="topbar__icon-btn topbar__collapse-btn"
      onClick={onCollapseClick}
      aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
    >
      <Icon name={collapsed ? 'chevronRight' : 'chevronLeft'} />
    </button>
    <span className="topbar__title">OmniAV</span>
  </header>
);

export default Topbar;
