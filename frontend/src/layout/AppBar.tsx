import * as React from 'react';
import { AppBar, Logout, UserMenu, useTranslate } from 'react-admin';
import {Link} from 'react-router-dom';
import GitHubIcon from '@mui/icons-material/GitHub';
import {
    Box,
    MenuItem,
    ListItemIcon,
    ListItemText,
    Typography,
    useMediaQuery,
    Theme,
} from '@mui/material';
import SettingsIcon from '@mui/icons-material/Settings';
import { useTheme } from '@mui/material/styles';
import Logo from './Logo';

const ConfigurationMenu = React.forwardRef((props, ref) => {
    const translate = useTranslate();
   
    return (
        <MenuItem
            component={Link}
            // @ts-ignore
            ref={ref}
            {...props}
            to="/configuration"
        >
            <ListItemIcon>
                <SettingsIcon />
            </ListItemIcon>
            <ListItemText>{translate('portal.configuration')}</ListItemText>
        </MenuItem>
    );
});
const CustomUserMenu = () => (
    <UserMenu>
        <ConfigurationMenu />
        <Logout />
    </UserMenu>
);

const CustomAppBar = (props: any) => {
    const isLargeEnough = useMediaQuery<Theme>(theme =>
        theme.breakpoints.up('sm')
    );
    const theme = useTheme();
    return (
        <AppBar
            {...props}
            color="secondary"
            elevation={1}
            userMenu={<CustomUserMenu />}
        >
            <Typography
                variant="h6"
                color="inherit"
                sx={{
                    flex: 1,
                    textOverflow: 'ellipsis',
                    whiteSpace: 'nowrap',
                    overflow: 'hidden',
                }}
                id="react-admin-title"
            />

            {isLargeEnough && <Logo />}
            {isLargeEnough && <Box component="span" sx={{ flex: 1 }} />}
            <a href="https://github.com/zebox/registry-admin"><GitHubIcon sx={{color:theme.palette.secondary.light}}/></a>
        </AppBar>
    );
};

export default CustomAppBar;
