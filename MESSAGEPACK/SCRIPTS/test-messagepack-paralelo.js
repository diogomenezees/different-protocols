import http from 'k6/http';
import { check } from 'k6';

const baseUrl = __ENV.BASE_URL || 'http://localhost:8090';

export let options = {
  vus: 50,           // Usuários virtuais simultâneos
  duration: '30s',   // Tempo total de execução
  thresholds: {
    // Limites e métricas para monitorar resultados
    'http_req_duration': ['avg<500', 'p(90)<1000', 'p(95)<1500'],  // exemplo limites
    'http_reqs': ['rate>100'], // taxa mínima desejada, pode ajustar
    'http_req_failed': ['rate<0.01'], // menos de 1% de erros
  },
  summaryTrendStats: ['avg', 'min', 'max', 'p(90)', 'p(95)'],  // Define quais stats são mostradas no resumo
};

export default function () {
  let id = Math.floor(Math.random() * 100) + 1;

  let res = http.get(`${baseUrl}/paralelo/nome-do-produto-${id}`);

  if (res.status !== 200) {
    console.error(`Status ${res.status} para produto`);
  }

  check(res, { 'Status 200': (r) => r.status === 200 });
}
