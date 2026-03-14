
const http = require('http');

async function test() {
    const projectName = 'project-1';
    
    console.log('Testing Create Chat...');
    const createRes = await new Promise((resolve, reject) => {
        const req = http.request({
            host: 'localhost',
            port: 3281,
            path: `/projects/${projectName}/chat/create`,
            method: 'POST',
            headers: {
                'HX-Request': 'true'
            }
        }, res => {
            console.log('Create Status:', res.statusCode);
            console.log('Create Headers:', JSON.stringify(res.headers, null, 2));
            let data = '';
            res.on('data', chunk => data += chunk);
            res.on('end', () => resolve({ status: res.statusCode, headers: res.headers, data }));
        });
        req.on('error', reject);
        req.end();
    });

    console.log('Create Body length:', createRes.data.length);
    if (createRes.data.length < 500) {
        console.log('Create Body:', createRes.data);
    } else {
        console.log('Create Body preview:', createRes.data.substring(0, 500));
    }

    // Try to extract chat ID from HX-Push-Url
    const pushUrl = createRes.headers['hx-push-url'];
    if (pushUrl) {
        const chatID = pushUrl.split('id=')[1];
        console.log('\nExtracted Chat ID:', chatID);

        console.log(`\nTesting Load Chat Messages for ${chatID}...`);
        const msgRes = await new Promise((resolve, reject) => {
            const req = http.request({
                host: 'localhost',
                port: 3281,
                path: `/projects/${projectName}/chat/messages/${chatID}`,
                method: 'GET',
                headers: {
                    'HX-Request': 'true'
                }
            }, res => {
                console.log('Messages Status:', res.statusCode);
                let data = '';
                res.on('data', chunk => data += chunk);
                res.on('end', () => resolve({ status: res.statusCode, data }));
            });
            req.on('error', reject);
            req.end();
        });
        console.log('Messages Body length:', msgRes.data.length);
        console.log('Messages Body preview:', msgRes.data.substring(0, 500));
    }
}

test().catch(console.error);
