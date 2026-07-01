export const APP_TIME_ZONE = "Asia/Jakarta";

const JAKARTA_OFFSET_MINUTES = 7 * 60;
const explicitTimezonePattern = /(Z|[+-]\d{2}:?\d{2})$/i;

const dateFormatter = new Intl.DateTimeFormat("id-ID", {
  day: "2-digit",
  month: "long",
  year: "numeric",
  timeZone: APP_TIME_ZONE,
});

const dateTimeFormatter = new Intl.DateTimeFormat("id-ID", {
  day: "2-digit",
  month: "long",
  year: "numeric",
  hour: "2-digit",
  minute: "2-digit",
  hour12: false,
  timeZone: APP_TIME_ZONE,
});

const fallback = "Tanggal tidak tersedia";
const timeFallback = "Waktu tidak tersedia";

export function formatDate(value?: string | Date | null) {
  const date = parseBackendTimestamp(value);
  return date ? dateFormatter.format(date) : fallback;
}

export function formatDateTime(value?: string | Date | null) {
  const date = parseBackendTimestamp(value);
  if (!date) return fallback;

  const parts = dateTimeFormatter.formatToParts(date);
  const day = getPart(parts, "day");
  const month = getPart(parts, "month");
  const year = getPart(parts, "year");
  const hour = getPart(parts, "hour");
  const minute = getPart(parts, "minute");

  if (!day || !month || !year || !hour || !minute) return fallback;
  return `${day} ${month} ${year}, ${hour}:${minute}`;
}

export function formatTime(value?: string | Date | null) {
  const date = parseBackendTimestamp(value);
  if (!date) return timeFallback;

  return new Intl.DateTimeFormat("id-ID", {
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
    timeZone: APP_TIME_ZONE,
  }).format(date);
}

export function formatShortDate(value?: string | Date | null) {
  const date = parseBackendTimestamp(value);
  if (!date) return fallback;

  return new Intl.DateTimeFormat("id-ID", {
    day: "2-digit",
    month: "short",
    year: "numeric",
    timeZone: APP_TIME_ZONE,
  }).format(date);
}

export function parseBackendDateOnly(value?: string | null) {
  const parts = dateOnlyParts(value);
  if (!parts) return null;

  const date = new Date(parts.year, parts.month - 1, parts.day);
  return Number.isNaN(date.getTime()) ? null : date;
}

export function formatDateOnly(value?: string | null) {
  const date = parseBackendDateOnly(value);
  if (!date) return fallback;

  return new Intl.DateTimeFormat("id-ID", {
    day: "2-digit",
    month: "short",
    year: "numeric",
  }).format(date);
}

export function todayDateOnly() {
  const parts = dateParts(new Date());
  if (!parts.year || !parts.month || !parts.day) return "";
  return `${parts.year}-${parts.month}-${parts.day}`;
}

export function compareDateOnly(
  left?: string | null,
  right?: string | null,
) {
  const leftKey = normalizeDateOnly(left);
  const rightKey = normalizeDateOnly(right);
  if (!leftKey && !rightKey) return 0;
  if (!leftKey) return 1;
  if (!rightKey) return -1;
  return leftKey.localeCompare(rightKey);
}

export function getTimeMinutes(value?: string | null) {
  const match = value?.trim().match(/^(\d{2}):(\d{2})/);
  if (!match) return 0;

  const [, hour, minute] = match;
  const hourValue = Number(hour);
  const minuteValue = Number(minute);
  if (
    Number.isNaN(hourValue) ||
    Number.isNaN(minuteValue) ||
    hourValue < 0 ||
    hourValue > 23 ||
    minuteValue < 0 ||
    minuteValue > 59
  ) {
    return 0;
  }

  return hourValue * 60 + minuteValue;
}

export function formatDateInputValue(value?: string | Date | null) {
  const date = parseBackendTimestamp(value);
  if (!date) return "";
  const parts = dateParts(date);
  if (!parts.year || !parts.month || !parts.day) return "";
  return `${parts.year}-${parts.month}-${parts.day}`;
}

export function formatTimeInputValue(value?: string | Date | null) {
  const date = parseBackendTimestamp(value);
  if (!date) return "";
  const parts = timeParts(date);
  if (!parts.hour || !parts.minute) return "";
  return `${parts.hour}:${parts.minute}`;
}

export function isSameDay(
  left?: string | Date | null,
  right?: string | Date | null,
) {
  const leftDate = parseBackendTimestamp(left);
  const rightDate = parseBackendTimestamp(right);
  if (!leftDate || !rightDate) return false;

  return dateKey(leftDate) === dateKey(rightDate);
}

export function isToday(value?: string | Date | null) {
  return isSameDay(value, new Date());
}

export function parseBackendTimestamp(value?: string | Date | null) {
  if (!value) return null;

  if (value instanceof Date) {
    return Number.isNaN(value.getTime()) ? null : value;
  }

  const trimmed = value.trim();
  if (!trimmed) return null;

  if (explicitTimezonePattern.test(trimmed)) {
    return parseNativeDate(trimmed);
  }

  // Format backend lama: DD-MM-YYYY HH:mm:ss
  const backendDateTime = trimmed.match(
    /^(\d{2})-(\d{2})-(\d{4})(?:\s+(\d{2}):(\d{2})(?::(\d{2}))?)?$/,
  );

  if (backendDateTime) {
    const [, day, month, year, hour = "00", minute = "00", second = "00"] =
      backendDateTime;

    return parseJakartaWallClock({
      year,
      month,
      day,
      hour,
      minute,
      second,
    });
  }

  // Format backend umum: YYYY-MM-DD atau YYYY-MM-DDTHH:mm:ss tanpa timezone.
  // Kolom timestamp tanpa timezone dari backend diperlakukan sebagai jam lokal Jakarta.
  const isoWithoutTimezone = trimmed.match(
    /^(\d{4})-(\d{2})-(\d{2})(?:[T\s](\d{2}):(\d{2})(?::(\d{2})(?:\.(\d{1,6}))?)?)?$/,
  );

  if (isoWithoutTimezone) {
    const [
      ,
      year,
      month,
      day,
      hour = "00",
      minute = "00",
      second = "00",
      fraction = "0",
    ] = isoWithoutTimezone;

    return parseJakartaWallClock({
      year,
      month,
      day,
      hour,
      minute,
      second,
      millisecond: fractionToMillisecond(fraction),
    });
  }

  return parseNativeDate(trimmed);
}

function getPart(
  parts: Intl.DateTimeFormatPart[],
  type: Intl.DateTimeFormatPartTypes,
) {
  return parts.find((part) => part.type === type)?.value;
}

function parseNativeDate(value: string) {
  const parsed = new Date(value);
  return Number.isNaN(parsed.getTime()) ? null : parsed;
}

function parseJakartaWallClock(parts: {
  year: string;
  month: string;
  day: string;
  hour?: string;
  minute?: string;
  second?: string;
  millisecond?: number;
}) {
  const parsed = new Date(
    Date.UTC(
      Number(parts.year),
      Number(parts.month) - 1,
      Number(parts.day),
      Number(parts.hour ?? "00"),
      Number(parts.minute ?? "00") - JAKARTA_OFFSET_MINUTES,
      Number(parts.second ?? "00"),
      parts.millisecond ?? 0,
    ),
  );

  return Number.isNaN(parsed.getTime()) ? null : parsed;
}

function fractionToMillisecond(value: string) {
  return Number(value.padEnd(3, "0").slice(0, 3));
}

function normalizeDateOnly(value?: string | null) {
  const parts = dateOnlyParts(value);
  if (!parts) return "";
  return `${parts.year}-${String(parts.month).padStart(2, "0")}-${String(
    parts.day,
  ).padStart(2, "0")}`;
}

function dateOnlyParts(value?: string | null) {
  const match = value?.trim().match(/^(\d{4})-(\d{2})-(\d{2})$/);
  if (!match) return null;

  const [, yearValue, monthValue, dayValue] = match;
  const year = Number(yearValue);
  const month = Number(monthValue);
  const day = Number(dayValue);
  if (
    Number.isNaN(year) ||
    Number.isNaN(month) ||
    Number.isNaN(day) ||
    month < 1 ||
    month > 12 ||
    day < 1 ||
    day > 31
  ) {
    return null;
  }

  return { year, month, day };
}

function dateKey(value: Date) {
  const parts = dateParts(value);
  return `${parts.year}-${parts.month}-${parts.day}`;
}

function dateParts(value: Date) {
  const parts = new Intl.DateTimeFormat("en-CA", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    timeZone: APP_TIME_ZONE,
  }).formatToParts(value);

  return {
    year: getPart(parts, "year"),
    month: getPart(parts, "month"),
    day: getPart(parts, "day"),
  };
}

function timeParts(value: Date) {
  const parts = new Intl.DateTimeFormat("en-GB", {
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
    timeZone: APP_TIME_ZONE,
  }).formatToParts(value);

  return {
    hour: getPart(parts, "hour"),
    minute: getPart(parts, "minute"),
  };
}
