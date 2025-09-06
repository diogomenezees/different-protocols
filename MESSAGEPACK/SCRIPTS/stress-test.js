import http from 'k6/http';
import { check } from 'k6';

// gera HTML bonitão no fim do teste
import { htmlReport } from 'https://raw.githubusercontent.com/benc-uk/k6-reporter/main/dist/bundle.js';

const baseUrl = __ENV.BASE_URL || 'http://localhost:8090';

export let options = {
  vus: 50,           // Usuários virtuais simultâneos
  duration: '1m',   // Tempo total de execução
  thresholds: {
    'http_req_duration': ['avg<500', 'p(90)<1000'],
    'http_reqs': ['rate>100'],
    'http_req_failed': ['rate<0.01'],
  },
  summaryTrendStats: ['avg', 'min', 'max', 'p(90)'],
};

export default function () {
  const id = Math.floor(Math.random() * 100) + 1;
  const res = http.get(`${baseUrl}/paralelo/nome-do-produto-${id}`);

  if (res.status !== 200) {
    console.error(`Status ${res.status} para produto`);
  }

  check(res, { 'Status 200': (r) => r.status === 200 });
}

// Salva HTML e JSON no volume montado em /output
export function handleSummary(data) {
  const base = '/output/resultado-json-X';

  return {
    [`${base}.page.html`]: htmlReport(data),
    [`${base}.summary.json`]: JSON.stringify(data, null, 2),
  };
}
