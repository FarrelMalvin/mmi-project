import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '1m', target: 100 },
    { duration: '28m', target: 100 },
    { duration: '1m', target: 0 },
  ],
};

// GANTI TOKEN INI dengan token yang didapat dari login manual/Postman
const AUTH_TOKEN = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJ5b3VyLWFwcC1uYW1lIiwic3ViIjoiMiIsImF1ZCI6WyJtYWluLWFwaSJdLCJleHAiOjE3NzcwMDAzNjAsIm5iZiI6MTc3Njk5NzY2MCwiaWF0IjoxNzc2OTk3NjYwLCJqdGkiOiI3OGZmYWYxOC1hNjg4LTRhMmItYjBhNy05ZTZiY2E0YjM3Y2YiLCJ1c2VyX2lkIjoyLCJqYWJhdGFuIjoiSFJHQSIsInRva2VuX3ZlcnNpb24iOjEsInRva2VuX3R5cGUiOiJhY2Nlc3MifQ.Z_QZuTO-wJA74vlOxplcxyyeyYhZCR2eSGMBieE8nUo';

export default function () {
  const baseUrl = 'http://localhost:8081/api/v1';
  
  const params = {
    headers: {
      'Authorization': `Bearer ${AUTH_TOKEN}`,
      'Content-Type': 'application/json',
    },
  };

  // --- 1. Testing PPD (Perjalanan Dinas) ---
  let resPPD = http.get(`${baseUrl}/ppd`, params);
  check(resPPD, { 'GET PPD status 200': (r) => r.status === 200 });

  let resPPDPending = http.get(`${baseUrl}/ppd/pending`, params);
  check(resPPDPending, { 'GET PPD Pending status 200': (r) => r.status === 200 });

  // --- 2. Testing RBS (Realisasi Bon) ---
  let resRBS = http.get(`${baseUrl}/rbs`, params);
  check(resRBS, { 'GET RBS status 200': (r) => r.status === 200 });

  let resRBSOptions = http.get(`${baseUrl}/rbs/options`, params);
  check(resRBSOptions, { 'GET RBS Options status 200': (r) => r.status === 200 });

  // --- 3. Testing User Profile ---
  let resProfile = http.get(`${baseUrl}/user/profile`, params);
  check(resProfile, { 'GET Profile status 200': (r) => r.status === 200 });

  sleep(1);
}