import { AppBar, Logout, UserMenu, ToggleThemeButton } from 'react-admin';
import { darkTheme, lightTheme } from './themes';
import {
    Box,
    Typography,
    useMediaQuery,
    Theme,
} from '@mui/material';
import Logo from './Logo';

const CustomUserMenu = () => (
    <UserMenu>
        <Logout />
    </UserMenu>
);

const CustomAppBar = (props: any) => {
    const isLargeEnough = useMediaQuery<Theme>(theme =>
        theme.breakpoints.up('sm')
    );
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
            <ToggleThemeButton
                lightTheme={lightTheme}
                darkTheme={darkTheme}
            />
        </AppBar>
    );
};

export default CustomAppBar;
