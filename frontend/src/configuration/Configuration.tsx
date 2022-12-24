import * as React from 'react';
import Card from '@mui/material/Card';
import Box from '@mui/material/Box';
import CardContent from '@mui/material/CardContent';
import Button from '@mui/material/Button';
import { useTranslate, useLocaleState, useTheme, Title, RaThemeOptions } from 'react-admin';

import { darkTheme, lightTheme } from '../layout/themes';

export const uiConfig = "current_ui_config";

export type UiConfig = {
    theme: string,
    language: string
}

const Configuration = () => {
    const translate = useTranslate();
    const [locale, setLocale] = useLocaleState();
    const [theme, setTheme] = useTheme();
    var config:UiConfig = {theme:"light", language:"en"};

    const themeSwitching = (themeValue: RaThemeOptions) => {

        config.theme= themeValue === darkTheme ? "dark" : "light";
        localStorage.setItem(uiConfig,JSON.stringify(config));
        setTheme(themeValue);
    }

    const languageSwitching = (language: string) => {
        config.language=language;
        localStorage.setItem(uiConfig,JSON.stringify(config));
        setLocale(language);
    }

    return (
        <Card>
            <Title title={translate('portal.configuration')} />
            <CardContent>
                <Box sx={{ width: '10em', display: 'inline-block' }}>
                    {translate('portal.theme.type')}
                </Box>
                <Button
                    variant="contained"
                    sx={{ margin: '1em' }}
                    color={
                        theme?.palette?.mode === 'light'
                            ? 'primary'
                            : 'secondary'
                    }
                    onClick={() => themeSwitching(lightTheme)}
                >
                    {translate('portal.theme.light')}
                </Button>
                <Button
                    variant="contained"
                    sx={{ margin: '1em' }}
                    color={
                        theme?.palette?.mode === 'dark'
                            ? 'primary'
                            : 'secondary'
                    }
                    onClick={() => themeSwitching(darkTheme)}
                >
                    {translate('portal.theme.dark')}
                </Button>
            </CardContent>
            <CardContent>
                <Box sx={{ width: '10em', display: 'inline-block' }}>
                    {translate('portal.language')}
                </Box>
                <Button
                    variant="contained"
                    sx={{ margin: '1em' }}
                    color={locale === 'en' ? 'primary' : 'secondary'}
                    onClick={() => languageSwitching('en')}
                >
                    en
                </Button>
                <Button
                    variant="contained"
                    sx={{ margin: '1em' }}
                    color={locale === 'ru' ? 'primary' : 'secondary'}
                    onClick={() => languageSwitching('ru')}
                >
                    ru
                </Button>
            </CardContent>
        </Card>
    );
};

export const SaveConfig = (option: RaThemeOptions | string) => {
    const [, setLocale] = useLocaleState();
    const [, setTheme] = useTheme();

    var config: UiConfig = {theme: "light", language: "en"};
    const configString = localStorage.getItem(uiConfig);

    if (configString !== null) {
        config = JSON.parse(configString);
    }

    if ((option as RaThemeOptions) !== undefined) {
        config.theme = option === darkTheme ? "dark" : "light";
        localStorage.setItem(uiConfig, JSON.stringify(config));
        setTheme(option as RaThemeOptions);
    }

    if (typeof option === 'string') {
        config.language = option;
        localStorage.setItem(uiConfig, JSON.stringify(config));
        setLocale(option);
    }

}

export default Configuration;
