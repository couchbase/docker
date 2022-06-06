const playwright = require('playwright');

(async () => {
    const browser = await playwright['chromium'].launch();
    const context = await browser.newContext();
    const page = await context.newPage();
    // Initial setup page
    await page.setViewportSize({
        width: 640,
        height: 480,
    });
    await page.goto('http://couchbase:8091/ui/index.html');
    await page.waitForSelector('button');
    await page.screenshot({
        path: `/output/setup-initial.jpg`,
        quality: 85,
        clip: {
            x: 0, y: 20,
            width: 640,
            height: 400,
        },
    });
    await page.click('text=Setup New Cluster')
    // Cluster creation
    await page.setViewportSize({
        width: 640,
        height: 480,
    });
    await page.waitForSelector('button');
    await page.fill('#for-cluster-name-field', 'cb-cluster')
    await page.click('#secure-password')
    await page.fill('#secure-password', 'AbCd123EfGh')
    await page.click('#secure-password-verify')
    await page.fill('#secure-password-verify', 'AbCd123EfGh')
    await page.screenshot({
        path: `/output/cluster-creation.jpg`,
        quality: 85,
        clip: {
            x: 0, y: 20,
            width: 640,
            height: 380,
        },
    });
    await page.click('text=Next: Accept Terms')
    // Terms & Conditions
    await page.setViewportSize({
        width: 640,
        height: 800,
    });
    await page.waitForSelector('css=[for="for-accept-terms"]', { waitFor: "visible" });
    await page.evaluate(() => {
        document.getElementById('for-accept-terms').click() // can't get this to work with page.click
    });
    await page.screenshot({
        path: `/output/finish-wizard.jpg`,
        quality: 85,
        clip: {
            x: 0, y: 20,
            width: 640,
            height: 600,
        },
    });
    await page.click('text=Finish With Defaults')
    // UI Home
    await page.setViewportSize({
        width: 1100,
        height: 400,
    });
    await page.waitForNavigation()
    await page.screenshot({
        path: `/output/ui-home.jpg`,
        quality: 85,
    });
    await page.click('text=Sample buckets');
    // Loading Sample Data
    await page.setViewportSize({
        width: 1100,
        height: 400,
    });
    await page.waitForNavigation()
    await page.click('css=[for="bucketbeer-sample"]')
    await page.screenshot({
        path: `/output/load-sample-data.jpg`,
        quality: 85,
    });
    await browser.close();
})();
