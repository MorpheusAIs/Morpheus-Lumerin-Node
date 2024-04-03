import Highcharts from 'highcharts';

/**
 * @param {[number, number][]} data
 */
export const renderChart = data => {
  const chart = {
    chart: {
      renderTo: 'container',
      type: 'spline'
    },
    title: {
      text: '',
      align: 'left'
    },

    legend: {
      // layout: 'vertical',
      align: 'right',
      verticalAlign: 'top',
      symbolRadius: 0,
      labelFormatter: function() {
        if (this.name === 'GH/s') {
          return 'GH/s';
        }
        return '';
      }
      // itemMarginTop: 10,
      // itemMarginBottom: 10
    },

    xAxis: {
      type: 'datetime',
      tickInterval: 1000 * 3600, // tick every hour
      labels: {
        formatter: function() {
          return Highcharts.dateFormat('%H %M', this.value);
        }
      }
    },
    yAxis: {
      min: 0
    },
    tooltip: {
      formatter: function() {
        return `${Highcharts.dateFormat(
          '%m/%d/%Y %H %M',
          this.x
        )} </br> Hashrate (5min): ${this.y} GH/s`;
      }
    },
    plotOptions: {
      series: {
        name: 'GH/s',
        pointInterval: 1000 * 60 * 5 // data every 5 minutes SET 5
      },
      spline: {
        lineWidth: 2,
        states: {
          hover: {
            lineWidth: 3
          }
        },
        marker: {
          enabled: false,
          radius: 2,
          states: {
            hover: {
              enabled: true,
              symbol: 'circle',
              radius: 2,
              lineWidth: 1
            }
          }
        }
      }
    },
    series: [
      {
        data: data
      }
    ]
  };
  return chart;
};
