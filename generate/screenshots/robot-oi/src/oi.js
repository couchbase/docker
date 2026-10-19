// Walks the Operational Insights setup wizard and console, capturing the
// screenshots embedded in the couchbase/operational-insights Docker Hub
// overview (generate/resources/operational-insights/README.md).
//
// Output filenames must stay in sync with the image URLs in that README.
const playwright = require('playwright');

const OI_HOST = process.env.OI_HOST || 'operational-insights';
const S3_ENDPOINT = process.env.S3_ENDPOINT || 'http://versitygw:7070';
const S3_BUCKET = process.env.S3_BUCKET || 'insights-storage';
const S3_ACCESS_KEY = process.env.S3_ACCESS_KEY || 'Administrator';
const S3_SECRET_KEY = process.env.S3_SECRET_KEY || 'password';
const CLUSTER_NAME = process.env.CLUSTER_NAME || 'insights-cluster';
const OI_USER = process.env.OI_USER || 'Administrator';
const OI_PASS = process.env.OI_PASS || 'password';
const OUT = process.env.OUTPUT_DIR || '/output';

const BASE = `http://${OI_HOST}:8091`;
const QUERY_URL = `http://${OI_HOST}:8095/analytics/service`;

// Keep the number of storage partitions low; the default of 128 is needless
// overhead for a single-node screenshot cluster.
const STORAGE_PARTITIONS = 16;

// Dimensions of the published images. The wizard pages are captured at a
// narrow width so they stay legible inline on Docker Hub; the console pages
// are captured wider so the left nav and results pane both fit.
const WIZARD_WIDTH = 640;
const CONSOLE = { width: 1024, height: 646 };

// Console navigation. Both the left nav and the section tab bar contain a link
// labelled "Workbench", so target them by route instead of by text.
const NAV_WORKBENCH = 'a[href^="#/cbas?"]';
const TAB_WORKBENCH = 'a[href^="#/cbas/workbench"]';
const TAB_SAMPLES = 'a[href^="#/cbas/samples"]';

const log = (msg) => console.log(`[oi-robot] ${msg}`);

// Set once the page exists, so a failure can be diagnosed from the output dir.
let failurePage = null;

async function shoot(page, name, size) {
  if (size) {
    await page.setViewportSize(size);
    // let the UI reflow at the new size before capturing
    await page.waitForTimeout(750);
  }
  await page.screenshot({ path: `${OUT}/${name}.png` });
  log(`captured ${name}.png`);
}

// Set an input's value the way Angular's change detection expects.
async function setField(page, id, value) {
  await page.evaluate(([id, value]) => {
    const el = document.getElementById(id);
    if (!el) throw new Error(`no element #${id}`);
    el.focus();
    el.value = value;
    el.dispatchEvent(new Event('input', { bubbles: true }));
    el.dispatchEvent(new Event('change', { bubbles: true }));
    el.blur();
  }, [id, value]);
}

async function clickById(page, id) {
  await page.evaluate((id) => {
    const el = document.getElementById(id);
    if (!el) throw new Error(`no element #${id}`);
    el.click();
  }, id);
}

async function setCheckbox(page, id, checked) {
  await page.evaluate(([id, checked]) => {
    const el = document.getElementById(id);
    if (!el) throw new Error(`no element #${id}`);
    if (el.checked !== checked) el.click();
  }, [id, checked]);
}

// Run a SQL++ statement against the Analytics service.
async function query(page, statement) {
  const res = await page.request.post(QUERY_URL, {
    form: { statement },
    headers: {
      Authorization:
        'Basic ' + Buffer.from(`${OI_USER}:${OI_PASS}`).toString('base64'),
    },
    timeout: 120000,
    failOnStatusCode: false,
  });
  if (!res.ok()) throw new Error(`query failed (${res.status()}): ${statement}`);
  return res.json();
}

// The Analytics service warms up for a while after the cluster is initialized;
// until it does, the Workbench renders an error banner instead of the editor.
async function waitForAnalytics(page, timeoutMs = 300000) {
  const deadline = Date.now() + timeoutMs;
  while (Date.now() < deadline) {
    try {
      await query(page, 'SELECT 1');
      return;
    } catch (e) {
      await page.waitForTimeout(5000);
    }
  }
  throw new Error('Analytics service did not become ready');
}

(async () => {
  const browser = await playwright.chromium.launch();
  const context = await browser.newContext();
  const page = await context.newPage();
  page.setDefaultTimeout(60000);
  failurePage = page;

  // ---------------------------------------------------------------- splash
  await page.setViewportSize({ width: WIZARD_WIDTH, height: 476 });
  await page.goto(`${BASE}/ui/index.html`);
  await page.waitForSelector('text=Setup New Cluster');
  await page.waitForTimeout(1000); // let the logo finish painting
  await shoot(page, 'setup-initial');

  // ------------------------------------------------- cluster + credentials
  await page.click('text=Setup New Cluster');
  await page.waitForSelector('#for-cluster-name-field');
  await setField(page, 'for-cluster-name-field', CLUSTER_NAME);
  await setField(page, 'secure-username', OI_USER);
  await setField(page, 'secure-password', OI_PASS);
  await setField(page, 'secure-password-verify', OI_PASS);
  await shoot(page, 'setup-wizard', { width: WIZARD_WIDTH, height: 448 });

  // ----------------------------------------------------------------- terms
  log('accepting terms');
  await page.click('button[type=submit]'); // Next: Accept Terms
  // The real checkbox input is visually hidden behind a styled label, so wait
  // for it to be attached rather than visible, and click it via the DOM.
  await page.waitForSelector('#for-accept-terms', { state: 'attached' });
  await clickById(page, 'for-accept-terms');
  await page.waitForTimeout(500);

  // ------------------------------------------------- disk / memory / blob
  log('opening disk/memory/blob configuration');
  await page.click('button[type=submit]'); // Configure Disk, Memory, Blob Storage
  await page.waitForSelector('#num_storage_partitions', { state: 'attached' });

  // S3-compatible, pointed at the Versity S3 Gateway service with its static
  // credentials and path-style addressing (the gateway is addressed by
  // container name, so virtual-host URLs cannot resolve).
  await clickById(page, 's3-compat');
  await page.waitForSelector('#bucket_endpoint', { state: 'attached' });
  await setField(page, 'bucket_endpoint', S3_ENDPOINT);
  await setField(page, 'bucket_name', S3_BUCKET);
  await setField(page, 'bucket_path_prefix', `${CLUSTER_NAME}/`);
  await setField(page, 'bucket_region', 'us-east-1');
  await clickById(page, 'cred-mode-static');
  await page.waitForSelector('#blob-access-key-id', { state: 'attached' });
  await setField(page, 'blob-access-key-id', S3_ACCESS_KEY);
  await setField(page, 'blob-secret-access-key', S3_SECRET_KEY);
  await setCheckbox(page, 'for-force-path-style', true);
  await setField(page, 'num_storage_partitions', String(STORAGE_PARTITIONS));
  // No screenshot of this form: the overview lists the values to enter in a
  // table instead, which readers can copy from.

  // -------------------------------------------------------------- finish up
  log('submitting cluster configuration');
  await page.click('text=Save & Finish');

  // The console loads (on the Dashboard) once the cluster is initialized.
  await page.waitForSelector(`${NAV_WORKBENCH}`, { timeout: 180000 });

  log('waiting for the Analytics service to warm up');
  await waitForAnalytics(page);
  await page.reload();
  await page.waitForSelector(`${NAV_WORKBENCH}`);
  await page.waitForTimeout(3000);

  // ------------------------------------------------------------- workbench
  // Captured before travel-sample is loaded, so the Databases pane shows only
  // the Default database.
  log('opening workbench');
  await page.setViewportSize(CONSOLE);
  await page.click(NAV_WORKBENCH);
  await page.waitForSelector('.wb-ace-editor', { state: 'attached' });
  await page.waitForTimeout(3000);
  await shoot(page, 'workbench');

  // ------------------------------------------------------- sample datasets
  log('opening samples');
  await page.click(TAB_SAMPLES);
  await page.waitForSelector('text=Load Sample Data');
  await page.waitForTimeout(1000);
  // Same hidden-input treatment as the terms checkbox.
  await page.evaluate(() => {
    const box = document.querySelector('input[type=checkbox]');
    if (box && !box.checked) box.click();
  });
  await page.waitForTimeout(500);
  await shoot(page, 'install-samples');

  log('loading travel-sample');
  await page.click('text=Load Sample Data');

  // Wait for ingestion to finish rather than guessing at a sleep.
  const deadline = Date.now() + 300000;
  let airlines = 0;
  while (Date.now() < deadline) {
    try {
      const r = await query(
        page,
        'SELECT VALUE COUNT(*) FROM `travel-sample`.`inventory`.`airline`'
      );
      airlines = (r.results && r.results[0]) || 0;
      if (airlines > 0) break;
    } catch (e) {
      // dataset not queryable yet
    }
    await page.waitForTimeout(5000);
  }
  if (!airlines) throw new Error('travel-sample did not finish loading');
  log(`travel-sample ready (${airlines} airlines)`);

  // ---------------------------------------------------------- sample query
  // Must stay in step with the query described in the README text.
  const statement = 'SELECT COUNT(*) FROM `travel-sample`.`inventory`.`airline`';

  log('running sample query');
  await page.click(TAB_WORKBENCH);
  await page.waitForSelector('.wb-ace-editor', { state: 'attached' });
  await page.waitForTimeout(2000);

  // The query editor is an ACE instance; drive it through ACE's own API rather
  // than synthesising keystrokes.
  await page.evaluate((statement) => {
    const el = document.querySelector('.wb-ace-editor');
    window.ace.edit(el).setValue(statement);
  }, statement);
  await page.waitForTimeout(500);

  await page.click('button:has-text("Execute")');
  await page.waitForSelector('text=success', { timeout: 120000 });
  await page.waitForTimeout(2500);
  await shoot(page, 'sample-query');

  await browser.close();
  log('done');
})().catch(async (err) => {
  console.error(`[oi-robot] FAILED: ${err.stack || err}`);
  if (failurePage) {
    try {
      await failurePage.screenshot({ path: `${OUT}/FAILURE.png`, fullPage: true });
      const text = await failurePage.evaluate(() => document.body.innerText);
      console.error(`[oi-robot] page text at failure:\n${text.slice(0, 1500)}`);
      console.error(`[oi-robot] wrote ${OUT}/FAILURE.png`);
    } catch (e) {
      console.error(`[oi-robot] could not capture failure state: ${e}`);
    }
  }
  process.exit(1);
});
