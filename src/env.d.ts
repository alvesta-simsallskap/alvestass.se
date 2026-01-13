type Runtime = import('@astrojs/cloudflare').Runtime<Env>;

declare namespace App {
  interface Locals extends Runtime {
    runtime: {
      env: Env;
    };
    otherLocals: {
      YOUR_ENV_KEY: string;
    };
  }
}