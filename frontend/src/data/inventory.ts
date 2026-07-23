import { EquipmentItem } from '../types/EquipmentItem';

const EQUIPMENT: EquipmentItem[] = [
  { id: 'bt-speaker', name: 'Bluetooth Speaker', available: 1 },
  { id: 'bt-speaker-mics', name: 'Bluetooth Speaker W/ 2 Wireless Mics', available: 1 },
  { id: 'webcam-mic', name: 'Computer Web Cam/Microphone', available: 1 },
  { id: 'cow-cart', name: 'COW (24 Laptops) Cart', available: 1 },
  { id: 'data-projector', name: 'Data Projector', available: 1 },
  { id: 'easel', name: 'Easel', available: 1 },
  { id: 'extra-anchovies', name: 'Extra Anchovies', available: 1 },
  { id: 'flip-chart', name: 'Flip Chart', available: 1 },
  { id: 'microphone', name: 'Microphone', available: 1 },
  { id: 'podium', name: 'Podium', available: 1 },
  { id: 'portable-podium-pa', name: 'Portable Podium & PA system', available: 1 },
  { id: 'projector-screen', name: 'Projector Screen - portable', available: 1 },
  { id: 'tripod', name: 'Tripod', available: 1 },
  { id: 'video-camera', name: 'Video Camera', available: 1 },
  { id: 'visualizer', name: 'Visualizer', available: 1 },
  { id: 'whiteboard', name: 'Whiteboard - portable', available: 1 },
  { id: 'zoom-equipment', name: 'Zoom Conferencing Equipment', available: 1 },
];

export const getEquipment = async (): Promise<EquipmentItem[]> => {
  return EQUIPMENT;
};
