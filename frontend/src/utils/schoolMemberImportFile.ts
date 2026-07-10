import * as XLSX from "xlsx";

export type ImportHeader = "fullName" | "email" | "role" | "classCode";

export const importTemplateRows = [
  ["fullName", "email", "role", "classCode"],
  ["Budi Santoso", "budi@siswa.sch.id", "student", "X-IPA-1"],
  ["Siti Rahma", "siti@guru.sch.id", "teacher", ""],
  ["Admin Sekolah", "admin@sekolah.sch.id", "admin", ""],
];

export function csvEscape(value: unknown) {
  const text = String(value ?? "");
  if (/[",\n\r]/.test(text)) {
    return `"${text.replace(/"/g, '""')}"`;
  }
  return text;
}

export function toCsv(rows: unknown[][]) {
  return rows.map((row) => row.map(csvEscape).join(",")).join("\n");
}

export function downloadTemplate() {
  const csv = toCsv(importTemplateRows);
  const blob = new Blob([csv], { type: "text/csv;charset=utf-8" });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = "template-import-warga-sekolah.csv";
  link.click();
  URL.revokeObjectURL(url);
}

export function downloadExcelTemplate() {
  const worksheet = XLSX.utils.aoa_to_sheet(importTemplateRows);
  worksheet["!cols"] = [{ wch: 24 }, { wch: 28 }, { wch: 14 }, { wch: 16 }];
  worksheet["!autofilter"] = { ref: "A1:D4" };

  const workbook = XLSX.utils.book_new();
  XLSX.utils.book_append_sheet(workbook, worksheet, "Import Warga");
  XLSX.writeFile(workbook, "template-import-warga-sekolah.xlsx", {
    compression: true,
  });
}

export function normalizeImportHeader(value: unknown): ImportHeader | "" {
  const normalized = String(value ?? "")
    .trim()
    .toLowerCase()
    .replace(/[\s_-]+/g, "");

  if (
    normalized === "fullname" ||
    normalized === "nama" ||
    normalized === "namalengkap"
  ) {
    return "fullName";
  }
  if (normalized === "email" || normalized === "alamatemail") {
    return "email";
  }
  if (normalized === "role" || normalized === "peran") {
    return "role";
  }
  if (normalized === "classcode" || normalized === "kodekelas") {
    return "classCode";
  }
  return "";
}

export function isExcelFile(file: File) {
  const name = file.name.toLowerCase();
  return (
    name.endsWith(".xlsx") ||
    file.type ===
      "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
  );
}

export async function convertXlsxToCsvFile(file: File) {
  const workbook = XLSX.read(await file.arrayBuffer(), {
    type: "array",
    cellDates: false,
  });
  const sheetName = workbook.SheetNames[0];
  if (!sheetName) {
    throw new Error("Workbook Excel tidak memiliki sheet.");
  }

  const worksheet = workbook.Sheets[sheetName];
  const rawRows = XLSX.utils.sheet_to_json<unknown[]>(worksheet, {
    header: 1,
    blankrows: false,
    defval: "",
  });

  const [rawHeader, ...dataRows] = rawRows;
  if (!rawHeader) {
    throw new Error("Sheet Excel kosong.");
  }

  const headers = rawHeader.map(normalizeImportHeader);
  const requiredHeaders: ImportHeader[] = ["fullName", "email", "role"];
  const missingHeaders = requiredHeaders.filter(
    (header) => !headers.includes(header),
  );
  if (missingHeaders.length > 0) {
    throw new Error("Header Excel wajib memuat fullName, email, dan role.");
  }

  const rows = dataRows
    .map((row) => {
      const mapped = new Map<string, unknown>();
      headers.forEach((header, index) => {
        if (header) mapped.set(header, row[index]);
      });
      return [
        mapped.get("fullName") ?? "",
        mapped.get("email") ?? "",
        mapped.get("role") ?? "",
        mapped.get("classCode") ?? "",
      ];
    })
    .filter((row) => row.some((value) => String(value ?? "").trim() !== ""));

  if (rows.length === 0) {
    throw new Error("Sheet Excel belum memiliki baris data.");
  }

  const csv = toCsv([["fullName", "email", "role", "classCode"], ...rows]);
  return new File([csv], file.name.replace(/\.xlsx$/i, ".csv"), {
    type: "text/csv;charset=utf-8",
  });
}
