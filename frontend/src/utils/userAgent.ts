// Small heuristic parser — not a full user-agent-parsing library (none was
// already a dependency, and adding one just for this felt heavier than
// warranted). Good enough for "which browser/device is this session on"
// without needing pixel-perfect version detection.
export function summarizeUserAgent(userAgent: string): string {
  if (!userAgent) return "Perangkat tidak diketahui";

  const browser = detectBrowser(userAgent);
  const os = detectOS(userAgent);

  if (browser && os) return `${browser} di ${os}`;
  if (browser) return browser;
  if (os) return `Perangkat ${os}`;
  return "Perangkat tidak diketahui";
}

function detectBrowser(ua: string): string {
  if (/Edg\//.test(ua)) return "Edge";
  if (/OPR\//.test(ua)) return "Opera";
  if (/CriOS\//.test(ua)) return "Chrome";
  if (/FxiOS\//.test(ua)) return "Firefox";
  if (/Chrome\//.test(ua) && !/Chromium/.test(ua)) return "Chrome";
  if (/Firefox\//.test(ua)) return "Firefox";
  if (/Safari\//.test(ua) && /Version\//.test(ua)) return "Safari";
  return "";
}

function detectOS(ua: string): string {
  if (/iPhone|iPad|iPod/.test(ua)) return "iOS";
  if (/Windows/.test(ua)) return "Windows";
  if (/Mac OS X/.test(ua)) return "macOS";
  if (/Android/.test(ua)) return "Android";
  if (/Linux/.test(ua)) return "Linux";
  return "";
}
