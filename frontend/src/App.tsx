import React from 'react';
import { Admin, CustomRoutes, memoryStore, Resource } from 'react-admin';
import {createBrowserHistory} from 'history';
import { Route } from 'react-router';
import polyglotI18nProvider from 'ra-i18n-polyglot';

import authProvider from './providers/authProviders';
import { Login, Layout } from './layout';
import AuthRequiredPage from './layout/AuthRequired';

import englishMessages from './i18n/en';
import { lightTheme } from './layout/themes';

import dataProvider from './providers/dataProvider';
// import './App.css';

import users from './users';

const history = createBrowserHistory();
const i18nProvider = polyglotI18nProvider(locale => {
  /*   if (locale === 'en') {
        return import('./i18n/en').then(messages => messages.default);
    }
   */
  // Always fallback on english
  return englishMessages;
}, 'en',{allowMissing: true});


function App() {
  return (
    <Admin
      title="Registry Admin Portal"
      dataProvider={dataProvider}
      authProvider={authProvider}
      store={memoryStore()}
      disableTelemetry
      loginPage={Login}
      // layout={Layout}
      i18nProvider={i18nProvider}
      theme={lightTheme}
      history={history}
    >

      <CustomRoutes>
        <Route path="/auth-required" element={<AuthRequiredPage />} />
      </CustomRoutes>
      <Resource name="users" {...users} />
    </Admin>
  );
}

export default App;
