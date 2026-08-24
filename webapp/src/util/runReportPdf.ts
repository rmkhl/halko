// Builds a per-run PDF report: header, one section per program step
// (stats + temperature chart), and the executed program as an appendix.
import { jsPDF } from "jspdf";
import autoTable from "jspdf-autotable";
import {
  Chart,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
} from "chart.js";
import { ExecutedProgram } from "../store/services/controlunitApi";
import { PowerSettings, RunNote, Step, StepType } from "../types/api";
import { LogRow, StepSegment, formatClock, parseExecutionLog, runStartedAt, segmentBySteps } from "./executionLog";

// Registered separately from ExecutionChart.tsx's identical call (Chart.js
// registration is idempotent) rather than importing "chart.js/auto": that
// convenience entry point is a thin re-export wrapper around this same
// "chart.js" module, and Parcel's production scope-hoisting/tree-shaking
// drops the wrapper's own module while leaving references to its exports
// object behind — a ReferenceError that only appears in production builds.
Chart.register(CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend);

export class NoStepDataError extends Error {
  constructor() {
    super("The execution log contains no program step data");
    this.name = "NoStepDataError";
  }
}

export interface RunReportInput {
  runName: string; // history entry name, e.g. "My program@2026-07-19T18:10:40+03:00"
  csv: string;
  executed?: ExecutedProgram; // undefined when the executed program could not be loaded
  notes?: RunNote[];
}

// Controlunit phases that are not program steps (used only when the
// executed program is unavailable and matching by step name is impossible).
const NON_STEP_PHASES = new Set(["Waiting", "Completed"]);

// The control unit synthesizes these in front of the authored program, so they
// appear in the stored program like any other step. They are not something the
// operator wrote, so the report shows them by name without a step number and
// starts the numbering at the first authored step.
const STARTUP_STEP_TYPES = new Set<StepType>(["equalize", "steam_prewarm"]);

const CHART_WIDTH_PX = 900;
const CHART_HEIGHT_PX = 400;

// Pairs each step with the number the report shows for it: authored steps count
// from 1, the control unit's synthesized startup steps get a dash.
const numberedProgramSteps = (steps: Step[]): [Step, string][] => {
  let number = 0;
  return steps.map((step) => {
    if (STARTUP_STEP_TYPES.has(step.type)) {
      return [step, "-"];
    }
    number += 1;
    return [step, String(number)];
  });
};

// A note is placed by time, not by the step name it carries: nothing stops a
// program from using the same step name twice, and matching on the name would
// misfile every note in such a run. The log's times are elapsed seconds from
// the run's start, so a note's absolute stamp maps into the same scale.
//
// When the run's start time is unknown the report falls back to an elapsed
// axis and so does this, matching on the step name instead. Each note prints
// the step it was stamped with, so a mismatch is visible rather than silent.
const notesInSegment = (
  notes: RunNote[],
  segment: StepSegment,
  startedAt?: number
): RunNote[] => {
  if (startedAt === undefined) {
    return notes.filter((note) => note.step === segment.step);
  }
  const first = segment.rows[0].time;
  const last = segment.rows[segment.rows.length - 1].time;
  return notes.filter((note) => {
    const elapsed = note.time - startedAt;
    return elapsed >= first && elapsed <= last;
  });
};

// Notes taken before the first step began (the preparation gap the header
// already reports) or after the last logged row. They are printed under the
// header rather than dropped.
const notesOutsideEverySegment = (
  notes: RunNote[],
  segments: StepSegment[],
  startedAt?: number
): RunNote[] => {
  const placed = new Set<RunNote>();
  segments.forEach((segment) => {
    notesInSegment(notes, segment, startedAt).forEach((note) => placed.add(note));
  });
  return notes.filter((note) => !placed.has(note));
};

// Every note prints its own step, even inside that step's section. It is what
// makes a misplaced note visible when the time fallback above had to match on
// the name. A note's time is an absolute stamp, so the clock is printable even
// when the run's start is unknown and the charts are on an elapsed axis.
const noteRows = (notes: RunNote[]): string[][] =>
  notes.map((note) => [
    formatClock(note.time),
    note.step,
    `${note.temperatures.kiln.toFixed(1)} / ${note.temperatures.material.toFixed(1)}`,
    note.text,
  ]);

const NOTE_COLUMNS = [["Time", "Step", "Kiln / material (°C)", "Note"]];

const formatDuration = (seconds: number): string => {
  const s = Math.max(0, Math.round(seconds));
  const h = Math.floor(s / 3600);
  const m = Math.floor((s % 3600) / 60);
  const sec = s % 60;
  return h > 0 ? `${h}h ${m}m ${sec}s` : `${m}m ${sec}s`;
};

const formatTimestamp = (timestamp?: number): string => {
  if (!timestamp) return "N/A";
  return new Date(timestamp * 1000).toLocaleString();
};

const mean = (values: number[]): number =>
  values.length ? values.reduce((a, b) => a + b, 0) / values.length : 0;

const describePower = (settings?: PowerSettings): string => {
  if (!settings) return "-";
  if (settings.min_delta !== undefined && settings.max_delta !== undefined) {
    return `delta ${settings.min_delta} to ${settings.max_delta} °C`;
  }
  if (settings.power !== undefined) {
    return `simple ${settings.power} %`;
  }
  return "-";
};

// Renders a temperature-only line chart of the segment into an offscreen
// canvas and returns it as a PNG data URL. Without a start time for the run
// the X-axis falls back to minutes elapsed since it started.
const renderSegmentChart = (rows: LogRow[], startedAt?: number): string => {
  const canvas = document.createElement("canvas");
  canvas.width = CHART_WIDTH_PX;
  canvas.height = CHART_HEIGHT_PX;
  const chart = new Chart(canvas, {
    type: "line",
    data: {
      labels: rows.map((row) =>
        startedAt ? formatClock(startedAt + row.time) : (row.time / 60).toFixed(1)
      ),
      datasets: [
        {
          label: "Material Temperature (°C)",
          data: rows.map((row) => row.material),
          borderColor: "rgb(255, 99, 132)",
          backgroundColor: "rgba(255, 99, 132, 0.5)",
          pointRadius: 2,
          tension: 0.3,
        },
        ...(rows.some((row) => row.kilnPrimary !== undefined)
          ? [
              {
                label: "Kiln 1 Temperature (°C)",
                data: rows.map((row) => row.kilnPrimary ?? null),
                borderColor: "rgb(255, 159, 64)",
                backgroundColor: "rgba(255, 159, 64, 0.5)",
                pointRadius: 2,
                tension: 0.3,
              },
              {
                label: "Kiln 2 Temperature (°C)",
                data: rows.map((row) => row.kilnSecondary ?? null),
                borderColor: "rgb(153, 102, 255)",
                backgroundColor: "rgba(153, 102, 255, 0.5)",
                pointRadius: 2,
                tension: 0.3,
              },
            ]
          : [
              {
                label: "Kiln Temperature (°C)",
                data: rows.map((row) => row.kiln),
                borderColor: "rgb(255, 159, 64)",
                backgroundColor: "rgba(255, 159, 64, 0.5)",
                pointRadius: 2,
                tension: 0.3,
              },
            ]),
      ],
    },
    options: {
      responsive: false,
      animation: false,
      devicePixelRatio: 2,
      plugins: {
        legend: { position: "top" },
      },
      scales: {
        x: {
          title: { display: true, text: startedAt ? "Time" : "Time (minutes)" },
          ticks: { maxTicksLimit: 12 },
        },
        y: {
          title: { display: true, text: "Temperature (°C)" },
          grace: "5%",
        },
      },
    },
  });
  const image = canvas.toDataURL("image/png");
  chart.destroy();
  return image;
};

type WithAutoTable = jsPDF & { lastAutoTable?: { finalY: number } };
const lastAutoTableY = (doc: jsPDF): number =>
  (doc as WithAutoTable).lastAutoTable?.finalY ?? 0;

export const generateRunReportPdf = (input: RunReportInput): jsPDF => {
  const rows = parseExecutionLog(input.csv);
  const segments = segmentBySteps(rows);
  const program = input.executed?.program;
  const stepNames = new Set((program?.steps ?? []).map((step) => step.name));
  const isStepSegment = (segment: StepSegment): boolean =>
    program ? stepNames.has(segment.step) : !NON_STEP_PHASES.has(segment.step);
  const stepSegments = segments.filter(isStepSegment);

  if (rows.length === 0 || stepSegments.length === 0) {
    throw new NoStepDataError();
  }

  const doc = new jsPDF({ unit: "pt", format: "a4" });
  const pageWidth = doc.internal.pageSize.getWidth();
  const pageHeight = doc.internal.pageSize.getHeight();
  const margin = 40;
  const contentWidth = pageWidth - 2 * margin;
  const chartHeight = (contentWidth * CHART_HEIGHT_PX) / CHART_WIDTH_PX;
  let y = margin + 12;

  // Header
  doc.setFontSize(16);
  doc.text(input.runName.split("@")[0], margin, y);
  y += 10;

  const startedAt = runStartedAt(input.executed?.started_at, input.runName);
  const runStart = rows[0].time;
  const runEnd = rows[rows.length - 1].time;
  const preparationSeconds = stepSegments[0].rows[0].time - runStart;

  autoTable(doc, {
    startY: y,
    margin: { left: margin, right: margin },
    theme: "plain",
    styles: { fontSize: 10, cellPadding: 2 },
    body: [
      ["Run", input.runName],
      ...(program?.description ? [["Description", program.description]] : []),
      ["State", input.executed?.state ?? "unknown"],
      ["Started", formatTimestamp(input.executed?.started_at)],
      ["Ended", formatTimestamp(input.executed?.completed_at)],
      ["Total duration", formatDuration(runEnd - runStart)],
      ["Preparation (before first step)", formatDuration(preparationSeconds)],
    ],
  });
  y = lastAutoTableY(doc) + 20;

  const notes = input.notes ?? [];
  const strayNotes = notesOutsideEverySegment(notes, stepSegments, startedAt);
  if (strayNotes.length > 0) {
    doc.setFontSize(13);
    doc.text("Notes", margin, y);
    y += 8;
    autoTable(doc, {
      startY: y,
      margin: { left: margin, right: margin },
      theme: "grid",
      styles: { fontSize: 9, cellPadding: 3 },
      head: NOTE_COLUMNS,
      body: noteRows(strayNotes),
    });
    y = lastAutoTableY(doc) + 20;
  }

  // One section per program step. Startup steps are unnumbered, so the counter
  // advances only for authored ones.
  let stepNumber = 0;
  stepSegments.forEach((segment, index) => {
    // Rough estimate of the section height for the page-break decision
    const sectionEstimate = 24 + 60 + chartHeight;
    if (y + sectionEstimate > pageHeight - margin && index > 0) {
      doc.addPage();
      y = margin + 12;
    }

    const programStep = program?.steps.find((step) => step.name === segment.step);
    const isStartup = programStep ? STARTUP_STEP_TYPES.has(programStep.type) : false;
    if (!isStartup) {
      stepNumber += 1;
    }
    let heading = isStartup ? segment.step : `Step ${stepNumber}: ${segment.step}`;
    if (programStep) {
      heading += ` (${programStep.type}`;
      if (programStep.temperature_target) {
        heading += `, target ${programStep.temperature_target} °C`;
      }
      heading += ")";
    }
    doc.setFontSize(13);
    doc.text(heading, margin, y);
    y += 8;

    const first = segment.rows[0];
    const last = segment.rows[segment.rows.length - 1];
    // Same rule as the charts: a run that recorded both sensors is reported
    // per sensor, one that did not keeps its single kiln figures.
    const hasSensorPair = segment.rows.some((row) => row.kilnPrimary !== undefined);
    const kilns = segment.rows.map((row) => row.kiln);
    const primaries = segment.rows.map((row) => row.kilnPrimary ?? row.kiln);
    const secondaries = segment.rows.map((row) => row.kilnSecondary ?? row.kiln);
    const materials = segment.rows.map((row) => row.material);
    const fmt = (value: number) => value.toFixed(1);

    autoTable(doc, {
      startY: y,
      margin: { left: margin, right: margin },
      theme: "grid",
      styles: { fontSize: 9, cellPadding: 3 },
      head: [[
        "Duration",
        "Kiln start -> end (°C)",
        ...(hasSensorPair
          ? ["Kiln 1 min - max (°C)", "Kiln 2 min - max (°C)"]
          : ["Kiln min - max (°C)"]),
        "Material start -> end (°C)",
        "Material min - max (°C)",
        "Avg heater/fan/hum (%)",
      ]],
      body: [[
        formatDuration(last.time - first.time),
        `${fmt(first.kiln)} -> ${fmt(last.kiln)}`,
        ...(hasSensorPair
          ? [
              `${fmt(Math.min(...primaries))} - ${fmt(Math.max(...primaries))}`,
              `${fmt(Math.min(...secondaries))} - ${fmt(Math.max(...secondaries))}`,
            ]
          : [`${fmt(Math.min(...kilns))} - ${fmt(Math.max(...kilns))}`]),
        `${fmt(first.material)} -> ${fmt(last.material)}`,
        `${fmt(Math.min(...materials))} - ${fmt(Math.max(...materials))}`,
        [
          Math.round(mean(segment.rows.map((row) => row.heater))),
          Math.round(mean(segment.rows.map((row) => row.fan))),
          Math.round(mean(segment.rows.map((row) => row.steam))),
        ].join(" / "),
      ]],
    });
    y = lastAutoTableY(doc) + 10;

    if (y + chartHeight > pageHeight - margin) {
      doc.addPage();
      y = margin;
    }
    doc.addImage(renderSegmentChart(segment.rows, startedAt), "PNG", margin, y, contentWidth, chartHeight);
    y += chartHeight + 12;

    // The notes taken during this step, under the curve they are about.
    // autoTable paginates itself, so a long list needs no page-break maths.
    const segmentNotes = notesInSegment(notes, segment, startedAt);
    if (segmentNotes.length > 0) {
      autoTable(doc, {
        startY: y,
        margin: { left: margin, right: margin },
        theme: "grid",
        styles: { fontSize: 9, cellPadding: 3 },
        head: NOTE_COLUMNS,
        body: noteRows(segmentNotes),
      });
      y = lastAutoTableY(doc) + 16;
    } else {
      y += 16;
    }
  });

  // Appendix: the executed program
  doc.addPage();
  y = margin + 12;
  doc.setFontSize(13);
  doc.text("Appendix: Program", margin, y);
  y += 8;

  if (program) {
    autoTable(doc, {
      startY: y,
      margin: { left: margin, right: margin },
      theme: "grid",
      styles: { fontSize: 9, cellPadding: 3 },
      head: [["#", "Name", "Type", "Target (°C)", "Runtime", "Heater", "Fan", "Steam"]],
      body: numberedProgramSteps(program.steps).map(([step, number]) => [
        number,
        step.name,
        step.type,
        step.temperature_target ? String(step.temperature_target) : "-",
        step.runtime ?? "-",
        describePower(step.heater),
        describePower(step.fan),
        describePower(step.steam),
      ]),
    });
  } else {
    doc.setFontSize(10);
    doc.text(
      "Program details unavailable (failed to load the executed program).",
      margin,
      y + 14
    );
  }

  return doc;
};
