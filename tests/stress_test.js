import http from "k6/http";
import { check, sleep, group } from "k6";

export const options = {
  stages: [
    { duration: "30s", target: 5 },
    { duration: "2m", target: 5 },

    { duration: "30s", target: 10 },
    { duration: "3m", target: 10 },

    { duration: "30s", target: 25 },
    { duration: "3m", target: 25 },

    { duration: "30s", target: 0 },
  ],
  thresholds: {
    http_req_failed: ["rate<0.01"],

    http_req_duration: ["p(95)<1000", "p(99)<2000"],

    "http_req_duration{endpoint:ppd_list}": ["p(95)<1000"],
    "http_req_duration{endpoint:ppd_pending}": ["p(95)<1000"],
    "http_req_duration{endpoint:ppd_detail}": ["p(95)<1500"],

    "http_req_duration{endpoint:rbs_list}": ["p(95)<1000"],
    "http_req_duration{endpoint:rbs_options}": ["p(95)<1000"],
    "http_req_duration{endpoint:rbs_detail}": ["p(95)<1500"],

    "http_req_duration{endpoint:user_profile}": ["p(95)<700"],
  },
};

const BASE_URL = __ENV.BASE_URL || "http://localhost:8081/api/v1";
const TOKEN = __ENV.TOKEN;

if (!TOKEN) {
  throw new Error(
    "TOKEN env wajib diisi. Contoh: TOKEN='xxx' k6 run tests/stress_test.js"
  );
}

const params = {
  headers: {
    Authorization: `Bearer ${TOKEN}`,
    "Content-Type": "application/json",
  },
};

function randomInt(min, max) {
  return Math.floor(Math.random() * (max - min + 1)) + min;
}

export default function () {
  group("PPD", () => {
    const resPPD = http.get(`${BASE_URL}/ppd?page=1&limit=10`, {
      ...params,
      tags: { endpoint: "ppd_list" },
    });

    check(resPPD, {
      "GET /ppd status 200": (r) => r.status === 200,
    });

    const resPPDPending = http.get(
      `${BASE_URL}/ppd/pending?page=1&limit=10`,
      {
        ...params,
        tags: { endpoint: "ppd_pending" },
      }
    );

    check(resPPDPending, {
      "GET /ppd/pending status 200": (r) => r.status === 200,
    });

    const ppdID = randomInt(1, 2000);

    const resPPDDetail = http.get(`${BASE_URL}/ppd/${ppdID}`, {
      ...params,
      tags: { endpoint: "ppd_detail" },
    });

    check(resPPDDetail, {
      "GET /ppd/:id status 200": (r) => r.status === 200,
    });
  });

  group("RBS", () => {
    const resRBS = http.get(`${BASE_URL}/rbs?page=1&limit=10`, {
      ...params,
      tags: { endpoint: "rbs_list" },
    });

    check(resRBS, {
      "GET /rbs status 200": (r) => r.status === 200,
    });

    const resRBSOptions = http.get(`${BASE_URL}/rbs/options`, {
      ...params,
      tags: { endpoint: "rbs_options" },
    });

    check(resRBSOptions, {
      "GET /rbs/options status 200": (r) => r.status === 200,
    });

    const rbsID = randomInt(1, 1200);

    const resRBSDetail = http.get(`${BASE_URL}/rbs/${rbsID}`, {
      ...params,
      tags: { endpoint: "rbs_detail" },
    });

    check(resRBSDetail, {
      "GET /rbs/:id status 200": (r) => r.status === 200,
    });
  });

  group("User", () => {
    const resProfile = http.get(`${BASE_URL}/user/profile`, {
      ...params,
      tags: { endpoint: "user_profile" },
    });

    check(resProfile, {
      "GET /user/profile status 200": (r) => r.status === 200,
    });
  });

  sleep(1);
}
