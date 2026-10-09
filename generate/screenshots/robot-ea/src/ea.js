// Walks the Enterprise Analytics setup wizard and console, capturing the
// screenshots embedded in the couchbase/enterprise-analytics Docker Hub
// overview (generate/resources/enterprise-analytics/README.md).
//
// Output filenames must stay in sync with the image URLs in that README.
const playwright = require('playwright');

const EA_HOST = process.env.EA_HOST || 'enterprise-analytics';
const S3_ENDPOINT = process.env.S3_ENDPOINT || 'http://s3mock:9090';
const S3_BUCKET = process.env.S3_BUCKET || 'ea-storage';
const CLUSTER_NAME = process.env.CLUSTER_NAME || 'ea-cluster';
const EA_USER = process.env.EA_USER || 'Administrator';
const EA_PASS = process.env.EA_PASS || 'password';
const OUT = process.env.OUTPUT_DIR || '/output';

const BASE = `http://${EA_HOST}:8091`;
const QUERY_URL = `http://${EA_HOST}:8095/analytics/service`;

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

const log = (msg) => console.log(`[ea-robot] ${msg}`);

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

// The wizard form scrolls inside its own container rather than the window, so
// position it by scrolling that container. Pass an element id to bring to the
// top of the form, or null for the very top.
async function scrollFormTo(page, anchorId) {
  await page.evaluate((anchorId) => {
    const findScroller = (el) => {
      let p = el && el.parentElement;
      while (p) {
        const style = getComputedStyle(p);
        if (
          (style.overflowY === 'auto' || style.overflowY === 'scroll') &&
          p.scrollHeight > p.clientHeight
        ) {
          return p;
        }
        p = p.parentElement;
      }
      return document.scrollingElement;
    };

    const probe = document.getElementById('setup_hostname');
    const scroller = findScroller(probe);
    if (!scroller) return;

    if (!anchorId) {
      scroller.scrollTop = 0;
      return;
    }
    const anchor = document.getElementById(anchorId);
    if (!anchor) return;
    // Offset by the anchor's label so it is not clipped at the top edge.
    const label = document.querySelector(`label[for="${anchorId}"]`);
    const top = (label || anchor).getBoundingClientRect().top;
    scroller.scrollTop += top - scroller.getBoundingClientRect().top;
  }, anchorId);
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
        'Basic ' + Buffer.from(`${EA_USER}:${EA_PASS}`).toString('base64'),
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
  await setField(page, 'secure-username', EA_USER);
  await setField(page, 'secure-password', EA_PASS);
  await setField(page, 'secure-password-verify', EA_PASS);
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

  // S3-compatible, pointed at the S3Mock service, with anonymous auth and
  // path-style addressing (S3Mock supports neither IAM nor virtual-host URLs).
  await clickById(page, 's3-compat');
  await page.waitForSelector('#bucket_endpoint', { state: 'attached' });
  await setField(page, 'bucket_endpoint', S3_ENDPOINT);
  await setField(page, 'bucket_name', S3_BUCKET);
  await setField(page, 'bucket_path_prefix', `${CLUSTER_NAME}/`);
  await setField(page, 'bucket_region', 'us-east-1');
  await clickById(page, 'cred-mode-anonymous');
  await setCheckbox(page, 'for-force-path-style', true);
  await setField(page, 'num_storage_partitions', String(STORAGE_PARTITIONS));

  // The config form is taller than one screen, so it is published as two
  // images: the top of the form, then the remainder scrolled into view.
  await page.setViewportSize({ width: WIZARD_WIDTH, height: 784 });
  await page.waitForTimeout(750);
  await scrollFormTo(page, null);
  await page.waitForTimeout(500);
  await shoot(page, 'blob-storage-config-1');

  // Second image picks up from the bucket fields, so the authentication mode
  // and the local storage paths are both visible. Taller than the equivalent
  // 2.0 image because 2.2 added the endpoint certificate field and the
  // Advanced section to this form.
  await page.setViewportSize({ width: WIZARD_WIDTH, height: 960 });
  await page.waitForTimeout(750);
  await scrollFormTo(page, 'bucket_name');
  await page.waitForTimeout(500);
  await shoot(page, 'blob-storage-config-2');

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
  console.error(`[ea-robot] FAILED: ${err.stack || err}`);
  if (failurePage) {
    try {
      await failurePage.screenshot({ path: `${OUT}/FAILURE.png`, fullPage: true });
      const text = await failurePage.evaluate(() => document.body.innerText);
      console.error(`[ea-robot] page text at failure:\n${text.slice(0, 1500)}`);
      console.error(`[ea-robot] wrote ${OUT}/FAILURE.png`);
    } catch (e) {
      console.error(`[ea-robot] could not capture failure state: ${e}`);
    }
  }
  process.exit(1);
});
