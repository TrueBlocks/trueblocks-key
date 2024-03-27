import { ENV } from './env.js';

console.log({ ENV });

const baseUrls = new Map([
    ["staging", "https://key-staging.auth.us-east-1.amazoncognito.com/"],
    ["production", ""],
]);
const clientIds = new Map([
    ["staging", "4ruo8kesvp7jjrl5rtirr703e6"],
    ["production", ""],
]);
const redirectUrls = new Map([
    ["staging", "https://key-staging-dc-frontend.s3.amazonaws.com/frontend/dashboard/index.html"],
    ["production", ""],
]);

const loginUrl = new URL("/login", baseUrls.get(ENV));
addUrlParams(loginUrl.searchParams);

const logoutUrl = new URL("/logout", baseUrls.get(ENV));
addUrlParams(logoutUrl.searchParams);

const code = (new URLSearchParams(location.search)).get('code');

const rootElement = document.querySelector('#root');
const loadingElement = document.querySelector('[data-js=loading]');
const errorElement = document.querySelector('[data-js=error]');
const loginBoxElement = document.querySelector('[data-js=loginbox]');
const logoutElement = document.querySelector('[data-js=logout]');
const dashboardTemplate = document.querySelector('template#dashboard');

function addUrlParams(searchParams) {
    searchParams.set("client_id", clientIds.get(ENV));
    searchParams.set("response_type", "code");
    searchParams.set("scope", "email openid phone profile");
    searchParams.set("redirect_uri", redirectUrls.get(ENV));
}

(function () {
    logoutElement.href = logoutUrl;
    window.addEventListener('click', (event) => {
        if (!event.target.matches('button[data-js=copy]')) return;

        copy(event);
    });
})();

(async function () {
    try {
        const resp = await fetch(
            `https://6zvhq12qa3.execute-api.us-east-1.amazonaws.com/prod/auth?code=${code}`,
            {
                method: 'POST'
            }
        );

        if (!code) {
            window.location = loginUrl;
        }

        if (!resp.ok) {
            const text = await resp.text()
            console.error('Error fetching credentials', resp.status, text);

            showError(resp.status, text);
            return;
        }

        const json = await resp.json()
        fillDashboardTemplate({
            email: json.email,
            endpoint: json.endpoint,
            username: json.username,
        });
    } catch (error) {
        window.location = loginUrl;

        // showError(null, error);
    }
})();

function fillDashboardTemplate({ email, endpoint, username }) {
    const emailElement = loginBoxElement.querySelector('[data-insert=email]');
    const clone = dashboardTemplate.content.cloneNode(true);
    const endpointElement = clone.querySelector('[data-insert=endpoint]');
    const usernameElement = clone.querySelector('[data-insert=username]');

    emailElement.textContent = email;
    endpointElement.textContent = endpoint;
    usernameElement.textContent = username;

    rootElement.replaceChildren(clone);
    loginBoxElement.classList.remove('hidden');
    loadingElement.classList.add('hidden');
}

function copy(event) {
    const code = event.target.parentElement.querySelector('code');
    if (!code) {
        return
    }
    navigator.clipboard.writeText(code.textContent);
}

function showError(statusCode, text) {
    const [codeElement, textElement] = errorElement.querySelectorAll('[data-insert=statusCode], [data-insert=error]');
    codeElement.textContent = statusCode;
    textElement.textContent = text;
    errorElement.removeAttribute('hidden');
    loadingElement.classList.add('hidden');
}