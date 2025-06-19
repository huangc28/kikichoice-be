/**
 * Generates a URL-friendly slug from a string that preserves Unicode characters
 * including Chinese characters and emojis
 */
export const generateSlug = (text: string): string => {
  return text
    .trim()
    // Replace spaces and multiple whitespace with hyphens
    .replace(/\s+/g, "-")
    // Remove only problematic URL characters while preserving Unicode
    // Remove: < > " ` % { } | \ ^ [ ] ` and control characters
    .replace(/[<>"'`%{}|\\^[\]`\x00-\x1f\x7f-\x9f]/g, "")
    // Replace multiple consecutive hyphens with single hyphen
    .replace(/\-\-+/g, "-")
    // Remove leading/trailing hyphens
    .replace(/^-+|-+$/g, "");
};

/**
 * Generates a unique slug by appending a number if conflicts exist
 */
export const generateUniqueSlug = (
  baseSlug: string,
  existingSlugs: Set<string>,
): string => {
  let slug = baseSlug;
  let counter = 1;

  while (existingSlugs.has(slug)) {
    slug = `${baseSlug}-${counter}`;
    counter++;
  }

  return slug;
};

/**
 * Alternative ASCII-only slug generator for systems that require ASCII URLs
 */
export const generateAsciiSlug = (text: string): string => {
  return text
    .toLowerCase()
    .trim()
    // Replace spaces and multiple whitespace with hyphens
    .replace(/\s+/g, "-")
    // Remove special characters except hyphens and alphanumeric (ASCII only)
    .replace(/[^\w\-]+/g, "")
    // Replace multiple consecutive hyphens with single hyphen
    .replace(/\-\-+/g, "-")
    // Remove leading/trailing hyphens
    .replace(/^-+|-+$/g, "");
};
