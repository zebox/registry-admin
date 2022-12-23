import React from 'react';
import { Admin, CustomRoutes, memoryStore, Resource, TranslationMessages } from 'react-admin';
import { createBrowserHistory } from 'history';
import { Route } from 'react-router';
import polyglotI18nProvider from 'ra-i18n-polyglot';

import authProvider from './providers/authProviders';
import { Login, Layout } from './layout';
import Configuration, {UiConfig, uiConfig} from './configuration/Configuration';
import Footer from './components/Footer/Footer';
import englishMessages from './i18n/en';
import russianMessages from './i18n/ru';
import { lightTheme, darkTheme } from './layout/themes';


import dataProvider from './providers/dataProvider';

import users from './users';
import groups from './groups';
import access from './access';
import repository from './registry';

const history = createBrowserHistory();


interface ITranslation {
  [key: string]: TranslationMessages;
}

const messages:ITranslation = {
  ru: russianMessages,
  en: englishMessages,
};

const i18nProvider = polyglotI18nProvider(locale => {

  const configString = localStorage.getItem(uiConfig);

  if (configString===null) {
    return messages.en;
  }

    const config = JSON.parse(configString);
    const {language}: UiConfig = config;

    if (language !== "" && messages[language]) {
        return messages[language];
    }

    return messages[locale] ? messages[locale] : messages.en;

}, 'en', {allowMissing: true});


function App() {
  const configString = localStorage.getItem(uiConfig);
  var currentTheme:string='light';
  if (configString!==null) {
    const  config = JSON.parse(configString);
    const {theme}:any = config;
    currentTheme=theme;
  };

  return (
      <React.Fragment>
          <Admin
              title="Registry Admin Portal"
              dataProvider={dataProvider}
              authProvider={authProvider}
              store={memoryStore()}
              disableTelemetry
              loginPage={Login}
              layout={Layout}
              i18nProvider={i18nProvider}
              theme={currentTheme && currentTheme === "light" ? lightTheme : darkTheme}
              history={history}
          >
              <Resource name="registry/catalog" {...repository} />
              <Resource name="access" {...access} />
              <Resource name="users" {...users} />
              <Resource name="groups" {...groups} />

              <CustomRoutes>
                  <Route path="/configuration" element={<Configuration/>}/>
              </CustomRoutes>
          </Admin>
          <Footer/>
      </React.Fragment>
  );
}

export default App;
