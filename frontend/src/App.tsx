import React from 'react';
import { Admin, CustomRoutes, memoryStore, Resource } from 'react-admin';
import { createBrowserHistory } from 'history';
import { Route } from 'react-router';
import polyglotI18nProvider from 'ra-i18n-polyglot';

import authProvider from './providers/authProviders';
import { Login, Layout } from './layout';
import Configuration, { themeSettingKey } from './configuration/Configuration';

import englishMessages from './i18n/en';
import { lightTheme, darkTheme } from './layout/themes';

import dataProvider from './providers/dataProvider';
// import './App.css';

import users from './users';
import groups from './groups';
import access from './access';
import repository from './registry';

const history = createBrowserHistory();
const i18nProvider = polyglotI18nProvider(locale => {
  if (locale === 'ru') {
    return import('./i18n/ru').then(messages => messages.default);
  }

  // Always fallback on english
  return englishMessages;
}, 'en', { allowMissing: true });


function App() {
  return (
    <Admin
      title="Registry Admin Portal"
      dataProvider={dataProvider}
      authProvider={authProvider}
      store={memoryStore()}
      disableTelemetry
      loginPage={Login}
      layout={Layout}
      i18nProvider={i18nProvider}
      theme={localStorage.getItem(themeSettingKey) === "light" ? lightTheme : darkTheme}
      history={history}
    >

      <CustomRoutes>
        <Route path="/configuration" element={<Configuration />} />

      </CustomRoutes>
      <Resource name="users" {...users} />
      <Resource name="groups" {...groups} />
      <Resource name="access" {...access} />
      <Resource name="registry/catalog" {...repository} />
    </Admin>
  );
}

export default App;
