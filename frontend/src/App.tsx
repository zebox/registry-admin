import React from 'react';
import { Admin, localStorageStore, Resource, resolveBrowserLocale, TranslationMessages } from 'react-admin';
import { createBrowserHistory } from 'history';
import polyglotI18nProvider from 'ra-i18n-polyglot';

import authProvider from './providers/authProviders';
import { Login, Layout } from './layout';
import Footer from './components/Footer/Footer';
import englishMessages from './i18n/en';
import russianMessages from './i18n/ru';
import { lightTheme } from './layout/themes';


import dataProvider from './providers/dataProvider';

import users from './users';
import groups from './groups';
import access from './access';
import repository from './registry';

const history = createBrowserHistory();
const STORE_VERSION = "2";

interface ITranslation {
  [key: string]: TranslationMessages;
}

const messages: ITranslation = {
  ru: russianMessages,
  en: englishMessages,
};

const i18nProvider = polyglotI18nProvider(
  locale => messages[locale] ? messages[locale] : messages.en,
  resolveBrowserLocale(),

  // other language should defined here
  [
    { locale: 'en', name: 'English' },
    { locale: 'ru', name: 'Русский' }
  ],

  { allowMissing: true }
);


function App() {
  return (
    <React.Fragment>
      <Admin
        title="Registry Admin Portal"
        dataProvider={dataProvider}
        authProvider={authProvider}
        store={localStorageStore(STORE_VERSION)}
        disableTelemetry
        loginPage={Login}
        layout={Layout}
        i18nProvider={i18nProvider}
        theme={lightTheme}
        history={history}
      >
        <Resource name="registry/catalog" {...repository} />
        <Resource name="access" {...access} />
        <Resource name="users" {...users} />
        <Resource name="groups" {...groups} />
      </Admin>
      <Footer />
    </React.Fragment>
  );
}

export default App;
