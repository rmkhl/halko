import React from "react";
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  ChartOptions,
} from "chart.js";
import { Line } from "react-chartjs-2";
import { Box, Paper, Typography, CircularProgress } from "@mui/material";

// Register Chart.js components
ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend
);

interface ExecutionChartProps {
  csvData?: string;
  title?: string;
  isLoading?: boolean;
}

// Static CSV data from ,run.csv (used as fallback)
const defaultCsvData = `time,step,steptime,material,oven,heater,fan,humidifier
6,Initializing,0,0.000000,0.000000,0,0,0
24,Initial Heating,0,20.000000,20.000000,0,0,0
84,Initial Heating,60,20.240005,22.500010,100,100,100
144,Initial Heating,120,21.920044,33.470104,0,100,0
204,Initial Heating,180,22.760063,32.630245,0,100,0
264,Initial Heating,240,24.440102,30.950327,0,100,0
324,Initial Heating,300,25.280121,30.110308,0,100,0
384,Initial Heating,360,26.120140,30.700296,100,100,100
444,Initial Heating,420,27.800179,41.890228,0,100,0
504,Initial Heating,480,28.640198,41.050369,0,100,0
564,Initial Heating,0,30.320236,39.370651,0,100,0
570,Secondary Heating,0,30.320236,39.370651,0,100,0
630,Secondary Heating,60,31.160255,38.530792,0,100,0
690,Secondary Heating,120,32.840134,36.851074,0,100,0
750,Secondary Heating,180,33.679993,41.401058,100,100,100
810,Secondary Heating,240,35.359711,43.131241,0,100,0
822,Acclimation Phase,0,35.359711,43.131241,0,100,0
882,Acclimation Phase,60,36.199570,42.291382,0,100,0
942,Acclimation Phase,120,37.879288,40.611664,0,100,0
1002,Acclimation Phase,180,38.719147,39.771805,0,100,0
1062,Acclimation Phase,240,40.398865,45.791862,0,100,0
1122,Acclimation Phase,0,41.238724,44.952003,0,100,0
1128,Cooling Phase,0,41.238724,44.952003,0,100,0
1188,Cooling Phase,60,42.918442,43.272285,0,100,0
1248,Cooling Phase,120,42.438522,42.432426,0,100,0
1308,Cooling Phase,180,41.598663,41.592567,0,100,0
1368,Cooling Phase,240,39.918945,39.912849,0,100,0
1428,Cooling Phase,300,39.079086,39.072990,0,100,0
1488,Cooling Phase,360,37.399368,37.393272,0,100,0
1548,Cooling Phase,420,36.559509,36.553413,0,100,0
1608,Cooling Phase,480,34.879791,34.873695,0,100,0
1668,Cooling Phase,540,34.039932,34.033836,0,100,0
1728,Cooling Phase,600,33.200073,33.193977,0,100,0
1788,Cooling Phase,660,31.520264,31.514166,0,100,0
1848,Cooling Phase,720,30.680244,30.674147,0,100,0
1908,Cooling Phase,780,29.000206,28.994108,0,100,0
1968,Cooling Phase,840,28.160187,28.154089,0,100,0
2028,Cooling Phase,900,26.480148,26.474051,0,100,0
2088,Cooling Phase,960,25.640129,25.634031,0,100,0
2148,Cooling Phase,1020,24.800110,24.794012,0,100,0
2208,Cooling Phase,1080,23.120071,23.113974,0,100,0
2268,Cooling Phase,1140,22.280052,22.273954,0,100,0
2328,Cooling Phase,1200,20.600014,20.593916,0,100,0
2376,Completed,0,20.000000,20.000000,0,100,0`;

interface DataPoint {
  time: number;
  step: string;
  steptime: number;
  material: number;
  oven: number;
  heater: number;
  fan: number;
  humidifier: number;
}

const parseCSV = (csv: string): DataPoint[] => {
  const lines = csv.trim().split("\n");
  const data: DataPoint[] = [];

  for (let i = 1; i < lines.length; i++) {
    const values = lines[i].split(",");
    data.push({
      time: parseFloat(values[0]),
      step: values[1],
      steptime: parseFloat(values[2]),
      material: parseFloat(values[3]),
      oven: parseFloat(values[4]),
      heater: parseFloat(values[5]),
      fan: parseFloat(values[6]),
      humidifier: parseFloat(values[7]),
    });
  }

  return data;
};

export const ExecutionChart: React.FC<ExecutionChartProps> = ({ csvData, title = "Program Execution Data", isLoading = false }) => {
  const dataPoints = parseCSV(csvData || defaultCsvData);

  if (isLoading) {
    return (
      <Paper
        sx={{
          width: "100%",
          padding: 3,
          marginTop: 3,
          display: "flex",
          justifyContent: "center",
          alignItems: "center",
          minHeight: 500,
        }}
      >
        <CircularProgress />
      </Paper>
    );
  }

  if (!csvData) {
    return (
      <Paper
        sx={{
          width: "100%",
          padding: 3,
          marginTop: 3,
        }}
      >
        <Typography variant="h6" gutterBottom>
          {title}
        </Typography>
        <Typography variant="body2" color="text.secondary" sx={{ textAlign: "center", padding: 4 }}>
          Select a program from history to view its execution chart
        </Typography>
      </Paper>
    );
  }

  // Convert time from seconds to minutes for X-axis
  const timeLabels = dataPoints.map((point) =>
    (point.time / 60).toFixed(1)
  );

  const chartData = {
    labels: timeLabels,
    datasets: [
      {
        label: "Material Temperature (°C)",
        data: dataPoints.map((point) => point.material),
        borderColor: "rgb(255, 99, 132)",
        backgroundColor: "rgba(255, 99, 132, 0.5)",
        yAxisID: "y-temperature",
        pointRadius: 2,
        tension: 0.3,
      },
      {
        label: "Oven Temperature (°C)",
        data: dataPoints.map((point) => point.oven),
        borderColor: "rgb(255, 159, 64)",
        backgroundColor: "rgba(255, 159, 64, 0.5)",
        yAxisID: "y-temperature",
        pointRadius: 2,
        tension: 0.3,
      },
      {
        label: "Heater Power (%)",
        data: dataPoints.map((point) => point.heater),
        borderColor: "rgb(54, 162, 235)",
        backgroundColor: "rgba(54, 162, 235, 0.5)",
        yAxisID: "y-power",
        pointRadius: 2,
        tension: 0.3,
        borderDash: [5, 5],
      },
      {
        label: "Fan Power (%)",
        data: dataPoints.map((point) => point.fan),
        borderColor: "rgb(34, 139, 34)",
        backgroundColor: "rgba(34, 139, 34, 0.5)",
        yAxisID: "y-power",
        pointRadius: 2,
        tension: 0.3,
        borderDash: [5, 5],
      },
      {
        label: "Humidifier Power (%)",
        data: dataPoints.map((point) => point.humidifier),
        borderColor: "rgb(255, 206, 86)",
        backgroundColor: "rgba(255, 206, 86, 0.5)",
        yAxisID: "y-power",
        pointRadius: 2,
        tension: 0.3,
        borderDash: [5, 5],
      },
    ],
  };

  const options: ChartOptions<"line"> = {
    responsive: true,
    maintainAspectRatio: false,
    interaction: {
      mode: "index" as const,
      intersect: false,
    },
    plugins: {
      legend: {
        position: "top" as const,
      },
      title: {
        display: false,
      },
      tooltip: {
        callbacks: {
          title: (context) => {
            const index = context[0].dataIndex;
            const point = dataPoints[index];
            return `Time: ${(point.time / 60).toFixed(1)} min | Step: ${point.step}`;
          },
        },
      },
    },
    scales: {
      x: {
        display: true,
        title: {
          display: true,
          text: "Time (minutes)",
        },
        ticks: {
          maxTicksLimit: 15,
        },
      },
      "y-temperature": {
        type: "linear" as const,
        display: true,
        position: "left" as const,
        title: {
          display: true,
          text: "Temperature (°C)",
        },
        min: 0,
        max: 50,
      },
      "y-power": {
        type: "linear" as const,
        display: true,
        position: "right" as const,
        title: {
          display: true,
          text: "Power (%)",
        },
        min: 0,
        max: 100,
        grid: {
          drawOnChartArea: false,
        },
      },
    },
  };

  return (
    <Paper
      sx={{
        width: "100%",
        padding: 3,
        marginTop: 3,
      }}
    >
      <Typography variant="h6" gutterBottom>
        {title}
      </Typography>
      <Box sx={{ height: 500, width: "100%" }}>
        <Line options={options} data={chartData} />
      </Box>
    </Paper>
  );
};
