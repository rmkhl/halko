import i18n from "i18next";
import { initReactI18next } from "react-i18next";
import { en } from "./en";

export type Translation = typeof en;

i18n.use(initReactI18next).init<Translation>({
  debug: false,
  fallbackLng: "en",
  interpolation: {
    escapeValue: false,
  },
  resources: {
    en: {
      translation: en,
    },
  },
});

export default i18n;
