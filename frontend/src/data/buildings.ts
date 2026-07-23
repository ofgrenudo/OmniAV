import { Building } from '../types/Building';

const BUILDINGS: Building[] = [
  { id: 'awh', name: 'Anna Whitten Hall' },
  { id: 'cah', name: 'Culinary Allied Health' },
  { id: 'cnm', name: 'Center for New Media' },
  { id: 'fic', name: 'Food Innovation Center' },
  { id: 'kvm', name: 'Kalamazoo Valley Museum' },
  { id: 'grv', name: 'Groves' },
  { id: 'ttc', name: 'Texas Township Campus' },
  { id: 'cosbar', name: 'Cosmatology and Barbering School' },
];

export const getBuildings = async (): Promise<Building[]> => {
  return BUILDINGS;
};
