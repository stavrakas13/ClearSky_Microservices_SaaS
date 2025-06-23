// public/js/student/statistics.js
import { flash } from '../../script.js';
import { getAvailableStats, getDistributions } from '../../api/stats.js';

function renderStatsTable(stats, onSelect) {
  const tbody = document.querySelector('.panel table tbody');
  tbody.innerHTML = '';
  stats.forEach((row, idx) => {
    const tr = document.createElement('tr');
    tr.innerHTML = `
      <td>${row.classTitle || row.course_name || row.course || '-'}</td>
      <td>${row.declarationPeriod || row.exam_period || '-'}</td>
      <td>${row.initialSubmissionDate?.split('T')[0] || '-'}</td>
      <td>${row.finalSubmissionDate?.split('T')[0] || '-'}</td>
    `;
    tr.style.cursor = 'pointer';
    tr.addEventListener('click', () => onSelect(row));
    if (idx === 0) tr.style.background = '#e6e7ea';
    tbody.appendChild(tr);
  });
}

function renderCharts(data) {
  const chartsDiv = document.getElementById('stats-charts');
  chartsDiv.innerHTML = '';
  if (!data || Object.keys(data).length === 0) {
    chartsDiv.innerHTML = '<div>No statistics available.</div>';
    return;
  }
  Object.entries(data).forEach(([key, dist]) => {
    const chartDiv = document.createElement('div');
    chartDiv.id = `echart-${key}`;
    Object.assign(chartDiv.style, {
      width: '100%', height: '220px', minWidth: '180px',
      background: '#fff', border: '1px solid #d0d7de',
      borderRadius: '6px', marginBottom: '1rem'
    });
    chartsDiv.appendChild(chartDiv);

    const chart = echarts.init(chartDiv);
    chart.setOption({
      title:    { text: key.toUpperCase(), left: 'center', top: 8, textStyle: { fontSize: 14 } },
      tooltip:  {},
      xAxis:    { type: 'category', data: dist.categories, name: 'Score', nameLocation: 'middle', nameGap: 25 },
      yAxis:    { type: 'value',    name: 'Count', minInterval: 1 },
      series: [{
        type: 'bar',
        data: dist.data,
        itemStyle: { color: '#006dd0' }
      }]
    });
  });
}

window.addEventListener('DOMContentLoaded', async () => {
  try {
    const stats = await getAvailableStats();
    if (!Array.isArray(stats) || stats.length === 0) {
      flash('No statistics available.');
      return;
    }

    renderStatsTable(stats, async (row) => {
      const filters = {
        course:            row.course_id || row.course   || row.classTitle,
        declarationPeriod: row.declarationPeriod       || row.exam_period,
        classTitle:        row.classTitle              || row.course_name || row.course
      };

      try {
        const data = await getDistributions(filters);
        renderCharts(data);
      } catch (err) {
        flash('Failed to load statistics: ' + err.message);
      }
    });

    // auto-click first row
    document.querySelector('.panel table tbody tr')?.click();
    flash('Statistics loaded.');
  } catch (err) {
    flash(err.message);
  }
});
