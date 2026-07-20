import { IconName } from '../components/layout/icons';

export interface NavItem {
  label: string;
  path: string;
  icon: IconName;
}

export interface NavSection {
  title?: string;
  items: NavItem[];
}

export const navSections: NavSection[] = [
  {
    items: [
      { label: 'Dashboard', path: '/', icon: 'dashboard' },
      { label: 'New Request', path: '/requests/new', icon: 'plus' },
      { label: 'My Requests', path: '/requests/mine', icon: 'list' },
    ],
  },
  {
    title: 'Technician',
    items: [
      { label: 'All Requests', path: '/requests', icon: 'inbox' },
      { label: 'Inventory', path: '/inventory', icon: 'box' },
    ],
  },
  {
    title: 'Admin',
    items: [{ label: 'Buildings', path: '/buildings', icon: 'building' }],
  },
];
