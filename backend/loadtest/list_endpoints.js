import http from "k6/http";
import { check, group, sleep } from "k6";
import { Rate, Trend } from "k6/metrics";

// 运行示例：
//   k6 run backend/loadtest/list_endpoints.js
//   CONFIG_FILE=backend/loadtest/list_endpoints.local.json k6 run backend/loadtest/list_endpoints.js
//   APIKEY="jc_xxx" VUS=50 HOLD=3m k6 run backend/loadtest/list_endpoints.js
//   APIKEY="jc_xxx" VUS=200 SLEEP=0 EXPECT_429=true k6 run backend/loadtest/list_endpoints.js
//
// 配置文件：
//   默认读取 backend/loadtest/list_endpoints.config.json；文件不存在时使用默认值。
//   可用 CONFIG_FILE 指定其他 JSON 配置文件。
//   环境变量仍可临时覆盖同名配置项，优先级：环境变量 > 配置文件 > 默认值。
//
// 配置项说明：
//   baseURL / BASE_URL：目标站点根地址，默认 https://course.sjtu.plus。
//   apiKey / APIKEY：Bearer API key，请求会携带 Authorization: Bearer <apiKey>。
//   cookie / COOKIE：浏览器登录态 Cookie，与 apiKey 二选一。
//   vus / VUS：升压结束后的虚拟用户数，默认 20。
//   rampUp / RAMP_UP：从 0 升到 vus 的时间，默认 30s。
//   hold / HOLD：保持 vus 并发的时间，默认 1m。
//   rampDown / RAMP_DOWN：从 vus 降到 0 的时间，默认 20s。
//   pageSize / PAGE_SIZE：list 接口的 page_size 查询参数，默认 20。
//   pages / PAGES：循环访问前多少页，默认 5。
//   sleep / SLEEP：每轮课程、教师、点评请求结束后的等待秒数，默认 1。
//   timeout / TIMEOUT：单个请求超时时间，默认 10s。
//   expect429 / EXPECT_429：是否预期触发限流；测限流时设为 true，默认 false。

const defaultConfig = {
  baseURL: "https://course.sjtu.plus",
  apiKey: "",
  cookie: "",
  vus: 20,
  rampUp: "30s",
  hold: "1m",
  rampDown: "20s",
  pageSize: 20,
  pages: 5,
  sleep: 1,
  timeout: "10s",
  expect429: false,
};

const fileConfig = readConfigFile(envString("CONFIG_FILE", "backend/loadtest/list_endpoints.config.json"));
const config = normalizeConfig({ ...defaultConfig, ...fileConfig, ...envConfig() });

const endpointOK = new Rate("endpoint_ok");
const rateLimited = new Rate("rate_limited");
const courseListDuration = new Trend("course_list_duration", true);
const teacherListDuration = new Trend("teacher_list_duration", true);
const reviewListDuration = new Trend("review_list_duration", true);

export const options = {
  discardResponseBodies: true,
  scenarios: {
    list_endpoints: {
      executor: "ramping-vus",
      stages: [
        { duration: config.rampUp, target: config.vus },
        { duration: config.hold, target: config.vus },
        { duration: config.rampDown, target: 0 },
      ],
    },
  },
  thresholds: {
    endpoint_ok: ["rate>0.99"],
    http_req_duration: ["p(95)<1000"],
    http_req_failed: [config.expect429 ? "rate<1.01" : "rate<0.01"],
    rate_limited: [config.expect429 ? "rate>0" : "rate<0.01"],
  },
};

export function setup() {
  const authModeCount = Number(Boolean(config.apiKey)) + Number(Boolean(config.cookie));
  if (authModeCount === 0) {
    throw new Error("set APIKEY or COOKIE for authenticated requests");
  }
  if (authModeCount > 1) {
    throw new Error("set only one auth mode: APIKEY or COOKIE");
  }

  return {
    requestParams: {
      headers: buildAuthHeaders(),
      timeout: config.timeout,
    },
  };
}

export default function (data) {
  const page = 1 + (__ITER % config.pages);
  const query = `page=${page}&page_size=${config.pageSize}`;

  group("course list", () => {
    const res = get(data.requestParams, "/api/course/", query, "course");
    courseListDuration.add(res.timings.duration);
    recordResult(res, "course list");
  });

  group("teacher list", () => {
    const res = get(data.requestParams, "/api/teacher/", query, "teacher");
    teacherListDuration.add(res.timings.duration);
    recordResult(res, "teacher list");
  });

  group("review list", () => {
    const res = get(data.requestParams, "/api/review", query, "review");
    reviewListDuration.add(res.timings.duration);
    recordResult(res, "review list");
  });

  sleep(config.sleep);
}

function get(params, path, query, endpoint) {
  return http.get(`${config.baseURL}${path}?${query}`, tagged(params, endpoint));
}

function buildAuthHeaders() {
  const headers = {
    Accept: "application/json",
    "Content-Type": "application/json",
  };
  if (config.apiKey) {
    headers.Authorization = `Bearer ${config.apiKey}`;
  }
  if (config.cookie) {
    headers.Cookie = normalizeCookie(config.cookie);
  }
  return headers;
}

function normalizeCookie(rawCookie) {
  return rawCookie
    .split(/,\s*/)
    .map((part) => part.split(";")[0])
    .filter(Boolean)
    .join("; ");
}

function tagged(params, endpoint) {
  return { ...params, tags: { ...params.tags, endpoint } };
}

function recordResult(res, name) {
  const ok = res.status === 200 || (config.expect429 && res.status === 429);
  endpointOK.add(ok);
  rateLimited.add(res.status === 429);
  check(res, {
    [`${name} status ok`]: () => ok,
  });
}

function envString(name, fallback) {
  const value = __ENV[name];
  return value === undefined || value === "" ? fallback : value;
}

function readConfigFile(path) {
  try {
    const content = open(path);
    if (!content.trim()) return {};
    return JSON.parse(content);
  } catch (err) {
    if (String(err).includes("no such file")) return {};
    throw new Error(`failed to read config file ${path}: ${err}`);
  }
}

function envConfig() {
  return compact({
    baseURL: __ENV.BASE_URL,
    apiKey: __ENV.APIKEY,
    cookie: __ENV.COOKIE,
    vus: parsePositiveNumber(__ENV.VUS),
    rampUp: __ENV.RAMP_UP,
    hold: __ENV.HOLD,
    rampDown: __ENV.RAMP_DOWN,
    pageSize: parsePositiveNumber(__ENV.PAGE_SIZE),
    pages: parsePositiveNumber(__ENV.PAGES),
    sleep: parsePositiveNumber(__ENV.SLEEP),
    timeout: __ENV.TIMEOUT,
    expect429: parseBool(__ENV.EXPECT_429),
  });
}

function normalizeConfig(input) {
  return {
    ...input,
    baseURL: String(input.baseURL).replace(/\/$/, ""),
    vus: requirePositiveNumber("vus", input.vus),
    pageSize: requirePositiveNumber("pageSize", input.pageSize),
    pages: requirePositiveNumber("pages", input.pages),
    sleep: requirePositiveNumber("sleep", input.sleep),
    expect429: Boolean(input.expect429),
  };
}

function compact(input) {
  const output = {};
  for (const [key, value] of Object.entries(input)) {
    if (value !== undefined && value !== "") output[key] = value;
  }
  return output;
}

function parsePositiveNumber(value) {
  if (value === undefined || value === "") return undefined;
  const parsed = Number(value);
  if (!Number.isFinite(parsed) || parsed <= 0) {
    throw new Error(`expected positive number, got ${value}`);
  }
  return parsed;
}

function requirePositiveNumber(name, value) {
  const parsed = Number(value);
  if (!Number.isFinite(parsed) || parsed <= 0) {
    throw new Error(`${name} must be a positive number`);
  }
  return parsed;
}

function parseBool(value) {
  if (value === undefined || value === "") return undefined;
  return ["1", "true", "yes", "on"].includes(String(value).toLowerCase());
}
