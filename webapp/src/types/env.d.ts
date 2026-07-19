// Parcel statically replaces process.env.* at build time. Declare the global
// here (not in the modules that use it) so tsc accepts the references while
// Parcel still sees `process` as a free identifier and performs the
// substitution — an in-module `declare` suppresses it and ships a raw
// `process.env` reference to the browser.
declare const process: { env: Record<string, string | undefined> };
