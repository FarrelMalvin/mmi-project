export const getApiErrorMessage = (
  err: any,
  fallback = "Terjadi kesalahan"
) => {
  const data = err?.response?.data;

  if (!data) return err?.message || fallback;

  if (typeof data === "string") return data;

  if (data.message) return data.message;
  if (data.error) return data.error;
  if (data.detail) return data.detail;

  if (Array.isArray(data.errors) && data.errors.length > 0) {
    return data.errors
      .map((item: any) => {
        if (typeof item === "string") return item;
        return item.message || item.error || item.field || JSON.stringify(item);
      })
      .join(", ");
  }

  return fallback;
};