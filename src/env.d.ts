type Runtime = import('@astrojs/cloudflare').Runtime<Env>;

declare namespace App {
  interface Locals extends Runtime {
    runtime: {
      env: {
        MJ_APIKEY_PUBLIC: string;
        MJ_APIKEY_PRIVATE: string;
        TURNSTILE_SECRET_KEY: string;
      };
    }
  }
}