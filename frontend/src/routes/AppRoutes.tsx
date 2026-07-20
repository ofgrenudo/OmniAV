import React from 'react';
import { Routes, Route } from 'react-router-dom';
import AppLayout from '../components/layout/AppLayout';
import Dashboard from '../pages/Dashboard';
import NewRequest from '../pages/NewRequest';
import MyRequests from '../pages/MyRequests';
import AllRequests from '../pages/AllRequests';
import Inventory from '../pages/Inventory';
import Buildings from '../pages/Buildings';

const AppRoutes: React.FC = () => (
  <Routes>
    <Route element={<AppLayout />}>
      <Route path="/" element={<Dashboard />} />
      <Route path="/requests/new" element={<NewRequest />} />
      <Route path="/requests/mine" element={<MyRequests />} />
      <Route path="/requests" element={<AllRequests />} />
      <Route path="/inventory" element={<Inventory />} />
      <Route path="/buildings" element={<Buildings />} />
    </Route>
  </Routes>
);

export default AppRoutes;
