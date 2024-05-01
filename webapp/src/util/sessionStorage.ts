export const getJSONFromSessionStorage = <T>(key: string): T | undefined => {
  const item = sessionStorage.getItem(key);

  return item ? (JSON.parse(item) as T) : undefined;
};

export const setJSONToSessionStorage = (key: string, item: any) => {
  const obj = JSON.stringify(item);

  sessionStorage.setItem(key, obj);
};

export const removeFromSessionStorage = (key: string) => {
  sessionStorage.removeItem(key);
};
