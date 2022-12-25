import {useTheme} from '@mui/material/styles';
import GitHubIcon from '@mui/icons-material/GitHub';

const Footer = () => {
    const theme = useTheme();

    return <div style={{
        position: 'fixed', right: 0, bottom: 0, left: 0, zIndex: 100,
        padding: 6,
        textAlign: 'left',
        width: 'fit-content'
    }}>
        <a href="https://github.com/zebox/registry-admin" style={{padding: 2}}><GitHubIcon
            sx={{color: theme.palette.primary.light}}/></a>
    </ div>
}

export default Footer;