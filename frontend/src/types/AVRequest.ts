import { WhenWhereData } from './WhenWhereData';
import { SelectedEquipment } from './SelectedEquipment';

export interface AVRequest {
  id: string;
  whenWhere: WhenWhereData;
  equipment: SelectedEquipment[];
  submittedAt: string;
}
