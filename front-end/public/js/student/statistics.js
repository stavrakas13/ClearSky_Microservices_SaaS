// public/js/student/statistics.js
import { flash } from '../../script.js';
import { getAvailableStats, getDistributions } from '../../api/stats.js';

function renderStatsTable(stats, onSelect, selectedIdx = 0) {
  const tbody = document.querySelector('.panel table tbody');
  tbody.innerHTML = '';
  stats.forEach((row, idx) => {
    const tr = document.createElement('tr');
    tr.innerHTML = `
      <td style="color:#006dd0;cursor:pointer;text-decoration:underline;">
        ${row.classTitle || row.course_name || row.course || '-'}
      </td>
      <td>${row.declarationPeriod || row.exam_period || '-'}</td>
      <td>${row.initialSubmissionDate?.split('T')[0] || '-'}</td>
      <td>${row.finalSubmissionDate?.split('T')[0] || '-'}</td>
    `;
    tr.style.cursor = 'pointer';
    if (idx === selectedIdx) {
      tr.style.background = '#e6e7ea';
      tr.classList.add('selected-course-row');
    }
    tr.addEventListener('click', () => onSelect(row, idx));
    tbody.appendChild(tr);
  });
}

function renderCharts(data, courseLabel, periodLabel) {
  const chartsDiv = document.getElementById('stats-charts');
  chartsDiv.innerHTML = '';

  // — Heading —
  const heading = document.createElement('div');
  heading.style.fontWeight   = 'bold';
  heading.style.fontSize     = '1.1rem';
  heading.style.marginBottom = '0.5rem';
  heading.textContent = 
    `Statistics for: ${courseLabel}` + 
    (periodLabel ? ` (${periodLabel})` : '');
  chartsDiv.appendChild(heading);

  if (!data || Object.keys(data).length === 0) {
    const msg = document.createElement('div');
    msg.textContent = 'No statistics available.';
    chartsDiv.appendChild(msg);
    return;
  }

  // — Grid wrapper —
  const grid = document.createElement('div');
  grid.className = 'grid-stats-inner';
  chartsDiv.appendChild(grid);

  // — Sort keys: grade, Q1–Q10, then anything else —
  const keys  = Object.keys(data);
  const qKeys = Array.from({ length: 10 }, (_, i) => `Q${i+1}`);
  const sortedKeys = [
    ...(['grade'].filter(k => keys.includes(k))),
    ...qKeys.filter(k => keys.includes(k)),
    ...keys.filter(k => k !== 'grade' && !qKeys.includes(k))
  ];

  sortedKeys.forEach(key => {
    const dist = data[key];
    if (!dist) return;

    // — Chart container —
    const chartDiv = document.createElement('div');
    chartDiv.id = `echart-${key}`;
    Object.assign(chartDiv.style, {
      height: '260px',
      background: '#fff',
      border: '1px solid #d0d7de',
      borderRadius: '10px',
      boxShadow: '0 2px 8px #e6e7ea'
    });
    grid.appendChild(chartDiv);

    const chart = echarts.init(chartDiv);
    chart.setOption({
      title:    { text: key.toUpperCase(), left: 'center', top: 8,
                  textStyle: { fontSize:15, fontWeight:600 } },
      tooltip:  { trigger: 'axis' },
      xAxis:    {
        type:        'category',
        data:        dist.categories,
        name:        'Score',
        nameLocation:'middle',
        nameGap:     25
      },
      yAxis:    { type:'value', name:'Count', minInterval:1 },
      series: [{
        type: 'bar',
        data: dist.data,
        itemStyle:{ color:'#006dd0', borderRadius:[4,4,0,0] }
      }]
    });
    window.addEventListener('resize', () => chart.resize());
  });
}

window.addEventListener('DOMContentLoaded', async () => {
  try {
    const stats = await getAvailableStats();
    if (!Array.isArray(stats) || !stats.length) {
      return flash('No statistics available.');
    }

    let selectedIdx = 0;
    const onSelect = async (row, idx) => {
      selectedIdx = idx;
      renderStatsTable(stats, onSelect, selectedIdx);

      const filters = {
        course:            row.course_id        || row.course   || row.classTitle,
        declarationPeriod: row.declarationPeriod || row.exam_period,
        classTitle:        row.classTitle       || row.course_name || row.course
      };

      try {
        const data = await getDistributions(filters);
        renderCharts(
          data,
          row.classTitle        || row.course_name || row.course || '-',
          row.declarationPeriod || row.exam_period || ''
        );
      } catch (err) {
        flash('Failed to load statistics: ' + err.message);
      }
    };

    renderStatsTable(stats, onSelect, selectedIdx);
    document.querySelector('.panel table tbody tr')?.click();
    flash('Statistics loaded.');
  } catch (err) {
    flash(err.message);
  }
});

// This file is already generic and works for all roles as long as the HTML structure matches.
