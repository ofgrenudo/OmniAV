import { Weekday } from './Weekday';

export interface WhenWhereData {
  requestName: string;
  firstDate: string;
  startTime: string;
  endTime: string;
  weeks: number;
  days: Weekday[];
  buildingId: string;
  roomNumber: string;
  comments: string;
}

export const createEmptyWhenWhere = (): WhenWhereData => ({
  requestName: '',
  firstDate: '',
  startTime: '',
  endTime: '',
  weeks: 1,
  days: [],
  buildingId: '',
  roomNumber: '',
  comments: '',
});
