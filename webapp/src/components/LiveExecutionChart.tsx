import React, { useEffect, useState, useRef, useMemo } from "react";
import { ExecutionChart } from "./ExecutionChart";
import { getApiEndpoints } from "../config/api";
import { useGetRunningProgramQuery } from "../store/services/controlunitApi";
import { useGetTemperaturesQuery } from "../store/services/sensorsApi";
import { RunningProgramResponse, TemperatureStatus, APIResponse, Step } from "../types/api";

interface LiveExecutionChartProps {
  title?: string;
}

export const LiveExecutionChart: React.FC<LiveExecutionChartProps> = ({
  title = "Live Program Execution"
}) => {
  const [csvData, setCsvData] = useState<string>("");
  const [isConnected, setIsConnected] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [noProgramRunning, setNoProgramRunning] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);
  const lastTimestampRef = useRef<number>(0);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const isMountedRef = useRef<boolean>(false);
  const isManualCloseRef = useRef<boolean>(false);
  const initialMaterialTempRef = useRef<number | null>(null);

  // Fetch running program and current temperatures
  const { data: runningProgramData } = useGetRunningProgramQuery(undefined, {
    pollingInterval: 5000,
    skipPollingIfUnfocused: true,
  });
  const { data: sensorData } = useGetTemperaturesQuery(undefined, {
    pollingInterval: 5000,
    skipPollingIfUnfocused: true,
  });

  // Calculate temperature range based on program targets and historical/current temps
  const temperatureRange = useMemo(() => {
    // Get program data to find highest target temperature
    const runningProgram = runningProgramData ? (runningProgramData as RunningProgramResponse) : undefined;
    const steps = runningProgram?.data?.program?.steps;

    let minTemp: number | null = null;

    // If we have historical CSV data, use the first data point for min temp
    if (csvData) {
      const lines = csvData.trim().split("\n");
      if (lines.length > 1) {
        const firstDataLine = lines[1]; // Skip header
        const values = firstDataLine.split(",");
        if (values.length >= 5) {
          const firstMaterial = parseFloat(values[3]);
          const firstOven = parseFloat(values[4]);
          if (!isNaN(firstMaterial) && !isNaN(firstOven)) {
            // Min is the lowest of: 0, first material temp, first oven temp
            minTemp = Math.floor(Math.min(0, firstMaterial, firstOven));
          }
        }
      }
    }

    // Fallback to current sensor data if no historical data
    if (minTemp === null) {
      const temperatures = sensorData ? (sensorData as APIResponse<Omit<TemperatureStatus, "delta">>) : undefined;
      const currentMaterialTemp = temperatures?.data?.material;

      // Store initial material temperature when first available
      if (currentMaterialTemp !== undefined && initialMaterialTempRef.current === null) {
        initialMaterialTempRef.current = currentMaterialTemp;
      }

      if (initialMaterialTempRef.current !== null) {
        minTemp = Math.floor(initialMaterialTempRef.current);
      }
    }

    // Calculate max from program steps
    if (minTemp !== null && steps && Array.isArray(steps)) {
      const maxTarget = Math.max(...steps.map((step: Step) => step.temperature_target || 0));

      if (maxTarget > 0) {
        const maxTemp = Math.ceil(maxTarget + 15);
        return { min: minTemp, max: maxTemp };
      }
    }

    return undefined;
  }, [sensorData, runningProgramData, csvData]);

  useEffect(() => {
    isMountedRef.current = true;

    const endpoints = getApiEndpoints();
    const baseUrl = endpoints.controlunit;
    const wsUrl = baseUrl.replace(/^http/, "ws") + "/engine/running/logws";

    // First, fetch accumulated data via REST
    const fetchAccumulatedData = async (): Promise<boolean> => {
      try {
        const response = await fetch(`${baseUrl}/engine/running/log`);

        if (response.status === 204) {
          // No program running
          setIsLoading(false);
          setNoProgramRunning(true);
          setCsvData("");
          return false;
        }

        if (response.ok) {
          const csvText = await response.text();
          setCsvData(csvText);
          setNoProgramRunning(false);

          // Extract last timestamp from the data
          const lines = csvText.trim().split("\n");
          if (lines.length > 1) {
            const lastLine = lines[lines.length - 1];
            const timestamp = parseInt(lastLine.split(",")[0], 10);
            if (!isNaN(timestamp)) {
              lastTimestampRef.current = timestamp;
            }
          }
          return true; // Program is running
        }
      } catch (error) {
        console.error("Failed to fetch accumulated log data:", error);
      } finally {
        setIsLoading(false);
      }
      return false; // No program running or error
    };

    const connectWebSocket = () => {
      // Don't connect if component is unmounted
      if (!isMountedRef.current) {
        return;
      }

      // Prevent duplicate connections - if one exists and is still connecting/open, don't create another
      if (wsRef.current && (wsRef.current.readyState === WebSocket.CONNECTING || wsRef.current.readyState === WebSocket.OPEN)) {
        return;
      }

      // Close existing one if it's closing/closed
      if (wsRef.current) {
        isManualCloseRef.current = true;
        wsRef.current.close();
        wsRef.current = null;
        isManualCloseRef.current = false;
      }

      // Clear any pending reconnect timeout
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
        reconnectTimeoutRef.current = null;
      }

      const ws = new WebSocket(wsUrl);
      wsRef.current = ws;

      ws.onopen = () => {
        setIsConnected(true);
      };

      ws.onmessage = (event) => {
        const message = event.data.trim();

        // Check if message is an error (not CSV data)
        if (message === "No program running") {
          setIsConnected(false);
          setCsvData("");
          setNoProgramRunning(true);
          // Close WebSocket and don't reconnect
          isManualCloseRef.current = true;
          ws.close();
          return;
        }

        // Skip empty messages
        if (!message) {
          return;
        }

        // Check if this is the header line
        if (message.startsWith("time,step,steptime")) {
          // Skip header - we already have it from REST or first WebSocket message
          return;
        }

        // Parse timestamp to check for duplicates
        const timestamp = parseInt(message.split(",")[0], 10);
        if (!isNaN(timestamp) && timestamp <= lastTimestampRef.current) {
          // Skip duplicate data
          return;
        }

        // Update last timestamp
        if (!isNaN(timestamp)) {
          lastTimestampRef.current = timestamp;
        }

        // Append new CSV line to existing data
        setCsvData((prevData) => {
          if (!prevData) {
            // First message should be header
            return message;
          }
          // Add newline and append new data
          const separator = prevData.endsWith("\n") ? "" : "\n";
          return prevData + separator + message;
        });
      };

      ws.onerror = () => {
        setIsConnected(false);
      };

      ws.onclose = () => {
        setIsConnected(false);
        // Only reconnect if it wasn't a manual close and component is still mounted
        if (!isManualCloseRef.current && isMountedRef.current && !noProgramRunning) {
          reconnectTimeoutRef.current = setTimeout(() => {
            // Re-fetch accumulated data on reconnect to catch up on any missed data points
            fetchAccumulatedData()
              .then((hasRunningProgram) => {
                if (hasRunningProgram) {
                  connectWebSocket();
                }
              })
              .catch(() => {
                // Error already logged, don't try to connect WebSocket
              });
          }, 5000);
        }
      };
    };

    // Fetch accumulated data first, then connect WebSocket only if program is running
    fetchAccumulatedData()
      .then((hasRunningProgram) => {
        if (hasRunningProgram) {
          connectWebSocket();
        }
      })
      .catch(() => {
        // Error already logged, don't connect WebSocket
      });

    // Cleanup on unmount
    return () => {
      isMountedRef.current = false;

      // Clear reconnect timeout
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
        reconnectTimeoutRef.current = null;
      }

      // Close WebSocket
      if (wsRef.current) {
        isManualCloseRef.current = true;
        wsRef.current.close();
        wsRef.current = null;
        isManualCloseRef.current = false;
      }
    };
  }, [noProgramRunning]);

  // Reset initial temperature when no program is running
  useEffect(() => {
    if (noProgramRunning) {
      initialMaterialTempRef.current = null;
    }
  }, [noProgramRunning]);

  if (noProgramRunning) {
    return (
      <div style={{ padding: "20px", textAlign: "center", color: "#666" }}>
        No program currently running
      </div>
    );
  }

  return (
    <ExecutionChart
      csvData={csvData || undefined}
      title={title}
      isLoading={isLoading || (!csvData && isConnected)}
      isLive={true}
      temperatureRange={temperatureRange}
    />
  );
}
