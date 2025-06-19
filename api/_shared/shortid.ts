import { randomBytes } from "crypto";

export const generateShortId = (length: number = 12): string => {
  return randomBytes(Math.ceil(length * 3 / 4))
    .toString("base64")
    .slice(0, length)
    .replace(/[+/-]/g, "_"); // Make URL-safe
};
