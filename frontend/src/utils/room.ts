// Room numbers are digits, optionally with an a/b/c wing suffix (e.g. "204", "123a"). Anything
// else — other letters, spaces, punctuation — isn't a room this system can route equipment to.
export const ROOM_NUMBER_PATTERN = /^[0-9a-cA-C]+$/;

export const ROOM_NUMBER_HINT = 'Numbers and the letters A, B, or C only (e.g. 204, 123a).';

export const isValidRoomNumber = (room: string): boolean => ROOM_NUMBER_PATTERN.test(room.trim());

/** Strips characters a room number can't contain, so typing simply can't produce an invalid value. */
export const sanitizeRoomNumber = (value: string): string => value.replace(/[^0-9a-cA-C]/g, '');
