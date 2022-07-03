import * as React from 'react';
import Card from '@mui/material/Card';
import Box from '@mui/material/Box';
import CardContent from '@mui/material/CardContent';
import Button from '@mui/material/Button';
import { useTranslate, useLocaleState, useTheme, Title, RaThemeOptions } from 'react-admin';

import { darkTheme, lightTheme } from '../layout/themes';

export const themeSettingKey = "RaCurrentTheme";

const Configuration = () => {
    const translate = useTranslate();
    const [locale, setLocale] = useLocaleState();
    const [theme, setTheme] = useTheme();

    const themeSwitching = (themeValue: RaThemeOptions) => {
        localStorage.setItem(themeSettingKey, themeValue === darkTheme ? "dark" : "light");
        setTheme(themeValue);
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
                    onClick={() => setLocale('en')}
                >
                    en
                </Button>
                <Button
                    variant="contained"
                    sx={{ margin: '1em' }}
                    color={locale === 'ru' ? 'primary' : 'secondary'}
                    onClick={() => setLocale('ru')}
                >
                    ru
                </Button>
            </CardContent>
        </Card>
    );
};

export default Configuration;
