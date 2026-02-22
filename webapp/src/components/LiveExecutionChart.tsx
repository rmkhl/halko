import React, { useEffect, useState, useRef } from "react";
import { ExecutionChart } from "./ExecutionChart";
import { getApiEndpoints } from "../config/api";

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
            console.log("Fetched accumulated log data:", lines.length, "lines, last timestamp:", timestamp);
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
        console.log("Skipping WebSocket connection - component unmounted");
        return;
      }

      // Prevent duplicate connections - if one exists and is still connecting/open, don't create another
      if (wsRef.current && (wsRef.current.readyState === WebSocket.CONNECTING || wsRef.current.readyState === WebSocket.OPEN)) {
        console.log("WebSocket already exists and is active, skipping duplicate connection");
        return;
      }

      // Close existing one if it's closing/closed
      if (wsRef.current) {
        console.log("Closing existing WebSocket before creating new one");
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

      console.log("Creating new WebSocket connection");
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
          console.log("Skipping duplicate timestamp:", timestamp);
          return;
        }

        // Update last timestamp
        if (!isNaN(timestamp)) {
          lastTimestampRef.current = timestamp;
        }

        console.log("Received new WebSocket data:", message);

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
          console.log("WebSocket closed unexpectedly, scheduling reconnect in 5 seconds");
          reconnectTimeoutRef.current = setTimeout(connectWebSocket, 5000);
        } else if (isManualCloseRef.current) {
          console.log("WebSocket closed manually, not reconnecting");
        } else {
          console.log("WebSocket closed, not reconnecting (component unmounted or no program running)");
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
      console.log("Cleaning up WebSocket on unmount");
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
    />
  );
}
