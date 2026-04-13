/**
 * Copyright 2026 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { state } from "../types";
import { setupMockDOM, createMockResponse } from "./helpers";
import { checkHealth, handleHealthSuccess, startHealthChecks, startRedirect, updateTimer, manualConnect, updateVisualProgress, updateDebugInfo, handleHealthTimeout } from "../health_module";
import { getGuacamoleUrl } from "../constants";
import { windowUtils } from "../window_utils";
import { setLanguage } from "../i18n_module";

(global as any).SVGCircleElement = class SVGCircleElement {
  style = { strokeDashoffset: '' };
};


describe("Health Module", () => {
  beforeEach(() => {
    setupMockDOM();
    state.isHealthy = false;
    state.pollCount = 0;
    state.startTime = Date.now();
    state.config = { ...state.config, timeoutMs: 5000, retryIntervalMs: 1000 };
    state.currentInterval = 1000;

    global.fetch = jest.fn() as unknown as typeof fetch;
    window.applyTranslations = jest.fn();
    window.updateUIFromConfig = jest.fn();
    window.updateDebugInfo = jest.fn();
    jest.spyOn(windowUtils, 'assign').mockImplementation(() => {});
  });

  afterEach(() => {
    jest.restoreAllMocks();
  });

  test("checkHealth handles success and transitions state", async () => {
    (global.fetch as jest.MockedFunction<typeof fetch>).mockResolvedValue(createMockResponse({}, { 'x-app-status': 'READY' }));

    await checkHealth();

    expect(state.isHealthy).toBe(true);
    expect(state.pollCount).toBe(1);
    expect(state.lastStatus).toBe('READY');
  });

  test("handleHealthSuccess updates UI classes", () => {
    handleHealthSuccess();
    expect(state.isHealthy).toBe(true);
    const statusIcon = document.getElementById('status-icon');
    expect(statusIcon?.classList.contains('animate-ready')).toBe(true);
  });

  test("startHealthChecks initiates polling loop", () => {
    jest.useFakeTimers();
    startHealthChecks();
    expect(state.currentInterval).toBe(1000);
    jest.useRealTimers();
  });

  test("startRedirect assigns correct URL based on protocol", () => {
    const assignSpy = jest.spyOn(windowUtils, 'assign').mockImplementation(() => {});

    state.config = { ...state.config, connectionId: 'RDP' };
    startRedirect();
    expect(assignSpy).toHaveBeenCalledWith(getGuacamoleUrl('RDP'));

    state.config = { ...state.config, connectionId: 'SSH' };
    startRedirect();
    expect(assignSpy).toHaveBeenCalledWith(getGuacamoleUrl('SSH'));

    state.config = { ...state.config, connectionId: '' };
    startRedirect();
    expect(assignSpy).toHaveBeenCalledWith('/');
  });

  test("handleHealthSuccess configures SVG transition", () => {
    const svg = document.createElement('div');
    jest.spyOn(document, 'querySelector').mockImplementation((sel) => {
      if (sel === 'main svg') return svg;
      return null;
    });

    handleHealthSuccess();
    expect(svg.style.transition).toBe("opacity 500ms ease-out");
    expect(svg.style.opacity).toBe("0");
  });

  test("checkHealth handles backoff or retry interval properly", async () => {
    (global.fetch as jest.Mock).mockRejectedValue(new Error('fail'));

    state.startTime = Date.now() - 10000;
    state.config = { ...state.config, timeoutMs: 5000 };
    await checkHealth();

    state.startTime = Date.now();
    await checkHealth();
  });

  test("updateTimer returns early if healthy", () => {
    state.isHealthy = true;
    const startTimer = window.setTimeout;
    window.setTimeout = jest.fn() as any;
    updateTimer();
    expect(window.setTimeout).not.toHaveBeenCalled();
    window.setTimeout = startTimer;
  });

  test("manualConnect redirects if healthy", () => {
    state.isHealthy = true;
    const spy = jest.spyOn(windowUtils, 'assign');
    manualConnect();
    expect(spy).toHaveBeenCalled();
  });

  test("startHealthChecks clears interval", () => {
    state.checkInterval = 12345 as any;
    jest.spyOn(global, 'clearTimeout');
    startHealthChecks();
    expect(clearTimeout).toHaveBeenCalledWith(12345);
  });
    test("checkHealth timeout & simulate", async () => {
          state.startTime = Date.now() - 100000;
          state.config = { ...state.config, timeoutMs: 10 };
          state.simulateSec = 1;
          await checkHealth();

          const origNow = Date.now;
          state.simulateSec = 10;
          Date.now = jest.fn().mockReturnValue(state.startTime + 20000); // remaining <= 0 path
          await checkHealth();
          Date.now = origNow;
        });
    test("checkHealth simulate remaining > 0", async () => {
          const origNow = Date.now;
          state.simulateSec = 10;
          Date.now = jest.fn().mockReturnValue(state.startTime + 1000);
          await checkHealth();
          Date.now = origNow;
        });
    test("checkHealth success", async () => {
          window.fetch = jest.fn().mockResolvedValue({ ok: true, headers: new Headers({ 'X-App-Status': 'READY' }) });
          await checkHealth();
        });
    test("updateVisualProgress", () => {
          document.body.innerHTML = '<svg id="progress-ring-path"></svg>';
          const el = document.getElementById("progress-ring-path");
          if (el) {
              Object.setPrototypeOf(el, SVGCircleElement.prototype);
              Object.defineProperty(el, 'style', { value: { strokeDashoffset: '0' }, writable: true });
              updateVisualProgress();
          }
        });
    test("updateDebugInfo branches", () => {
          document.body.innerHTML = '<div id="debug-container"></div><div id="debug-info-content"></div>';
          state.latencyMs = null;
          state.lastStatus = "STARTING";
          Object.defineProperty(navigator, 'userAgent', { value: 'Safari', configurable: true });
          updateDebugInfo();
          Object.defineProperty(navigator, 'userAgent', { value: 'Firefox', configurable: true });
          updateDebugInfo();
          Object.defineProperty(navigator, 'userAgent', { value: 'Chrome', configurable: true });
          updateDebugInfo();
          state.isHealthy = true;
          updateDebugInfo();
        });
    test("handleHealthTimeout with el", () => {
          document.body.innerHTML = '<div id="status-message"></div>';
          handleHealthTimeout();
        });
    test("handleHealthSuccess with elements & no redirect", () => {
          document.body.innerHTML = '<div id="main-status"></div><div id="status-icon"></div><main><svg></svg></main><div class="starfield-container animate-drift"></div>';
          state.config = { ...state.config, autoRedirect: false };
          state.currentModal = "some";
          state.checkInterval = 123 as any;
          handleHealthSuccess();
        });
    test("startHealthChecks timeout backoff", async () => {
          jest.useFakeTimers();
          state.startTime = Date.now() - 100000;
          state.config = { ...state.config, timeoutMs: 10 };
          window.fetch = jest.fn().mockRejectedValue(new Error("err"));
          startHealthChecks();
          jest.advanceTimersByTime(10);

          state.isHealthy = true; // hit early return in poll
          jest.advanceTimersByTime(10000);
          jest.useRealTimers();
        });
    test("startHealthChecks normal backoff", async () => {
          jest.useFakeTimers();
          state.startTime = Date.now();
          state.config = { ...state.config, timeoutMs: 100000 };
          window.fetch = jest.fn().mockRejectedValue(new Error("err"));
          startHealthChecks();
          jest.advanceTimersByTime(10);
          jest.useRealTimers();
        });
    test("updateTimer branches", () => {
          state.startTime = Date.now() - 100000;
          state.config = { ...state.config, timeoutMs: 10 };
          state.simulateSec = 1;
          document.body.innerHTML = '<div id="live-timer"></div>';
          window.applyTranslations = jest.fn();
          updateTimer();
          expect(window.applyTranslations).toHaveBeenCalled();
        });
    test("manualConnect not healthy", () => {
          state.isHealthy = false;
          manualConnect();
        });
    test("startRedirect", () => {
          startRedirect();
          expect(windowUtils.assign).toHaveBeenCalled();
        });
    test("health_module updateVisualProgress branches", () => {
        jest.spyOn(window, 'requestAnimationFrame').mockImplementation(() => { return 1; });
        state.isHealthy = true;
        updateVisualProgress();

        state.isHealthy = false;
        document.body.innerHTML = '';
        updateVisualProgress();

        document.body.innerHTML = '<svg><circle id="progress-ring-path"></circle></svg>';
        const ring = document.getElementById('progress-ring-path') as unknown as SVGCircleElement;
        Object.defineProperty(ring, 'style', { value: { strokeDashoffset: '' }, writable: true });
        Object.setPrototypeOf(ring, (global as any).SVGCircleElement.prototype);
        updateVisualProgress();
      });
    test("health_module startHealthChecks and poll", async () => {
        jest.useFakeTimers();
        const clearTimeoutSpy = jest.spyOn(window, 'clearTimeout');
        jest.spyOn(windowUtils, 'assign').mockImplementation(() => {});

        // Stub fetch for checkHealth
        global.fetch = jest.fn(() => Promise.resolve({ ok: false })) as any;

        state.checkInterval = 12345 as any;
        state.config = { ...state.config, timeoutMs: -1000 };
        state.startTime = Date.now();
        startHealthChecks();

        // Let async promises flush and timers fire
        await Promise.resolve();
        jest.runOnlyPendingTimers();
        await Promise.resolve();
        jest.useRealTimers();
        clearTimeoutSpy.mockRestore();
      });
    test("health_module handleHealthSuccess branch", () => {
        jest.spyOn(windowUtils, 'assign').mockImplementation(() => {});
        state.isHealthy = true;
        handleHealthSuccess(); // should return early
        state.isHealthy = false;
        handleHealthSuccess();
      });
    test("updateDebugInfo missing window.t", () => {
        document.body.innerHTML = `
      <div id="debug-container"></div>
      <div id="debug-info-content"></div>
    `;
        window.t = undefined as any;
        updateDebugInfo();
      });
    test("updateDebugInfo missing contentEl", () => {
        document.body.innerHTML = '<div id="debug-container"></div>';
        updateDebugInfo();
      });
    test("updateDebugInfo status OFFLINE", () => {
        document.body.innerHTML = `
      <div id="debug-container"></div>
      <div id="debug-info-content"></div>
    `;
        state.lastStatus = "OFFLINE";
        updateDebugInfo();
      });
    test("health_module poll isHealthy becomes true", async () => {
        jest.useFakeTimers();
        global.fetch = jest.fn(() => Promise.resolve({ ok: true, headers: new Headers({ 'X-App-Status': 'READY' }) })) as any;
        state.isHealthy = false;
        state.config = { ...state.config, timeoutMs: 10000 };
        state.startTime = Date.now();
        startHealthChecks();
        await Promise.resolve();
        jest.runOnlyPendingTimers();
        await Promise.resolve();
        jest.useRealTimers();
      });
    test("startHealthChecks with falsy checkInterval", () => {
        state.checkInterval = 0 as any;
        startHealthChecks();
      });
    test("Requirement: Debug overlay labels update dynamically when language changes", async () => {
        state.config = { ...state.config, showDebug: true };
        state.translations['fr'] = {
          name: 'French',
          native: 'Français',
          dict: { 'label_debug_latency': 'Latence_FR' }
        };

        const { updateDebugInfo } = await import('../health_module');
        const { t } = await import('../i18n_module');
        window.t = jest.fn((k, def) => t(k, def));
        window.updateDebugInfo = jest.fn().mockImplementation(() => updateDebugInfo());

        await setLanguage('fr');

        const content = document.getElementById('debug-info-content')!;
        expect(content.innerHTML).toContain('Latence_FR');
      });
});


