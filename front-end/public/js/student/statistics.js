// student/statistics.js
import { flash } from '../../script.js';
import { getDistributions } from '../../api/stats.js';

window.addEventListener('DOMContentLoaded', async () => {
  try {
    const { data } = await getDistributions({ /* filters */ });
    // Remove the old message
    flash('Statistics loaded.');

    // Render ECharts for each distribution
    const chartsDiv = document.getElementById('stats-charts');
    chartsDiv.innerHTML = ''; // clear previous

    if (!data || Object.keys(data).length === 0) {
      chartsDiv.innerHTML = '<div>No statistics available.</div>';
      return;
    }

    Object.entries(data).forEach(([key, dist], idx) => {
      // Create a container for each chart
      const chartId = `echart-${key}`;
      const chartDiv = document.createElement('div');
      chartDiv.id = chartId;
      chartDiv.style.width = '100%';
      chartDiv.style.height = '220px';
      chartDiv.style.minWidth = '180px';
      chartDiv.style.background = '#fff';
      chartDiv.style.border = '1px solid #d0d7de';
      chartDiv.style.borderRadius = '6px';
      chartDiv.style.marginBottom = '1rem';
      chartsDiv.appendChild(chartDiv);

      // Render ECharts bar chart
      const chart = echarts.init(chartDiv);
      chart.setOption({
        title: { text: key.toUpperCase(), left: 'center', top: 8, textStyle: { fontSize: 14 } },
        tooltip: {},
        xAxis: {
          type: 'category',
          data: dist.labels, // changed from dist.categories
          name: 'Score',
          nameLocation: 'middle',
          nameGap: 25,
        },
        yAxis: {
          type: 'value',
          name: 'Count',
          minInterval: 1,
        },
        series: [{
          type: 'bar',
          data: dist.counts, // changed from dist.data
          itemStyle: { color: '#006dd0' }
        }]
      });
    });
  } catch (err) {
    flash(err.message);
  }
});
