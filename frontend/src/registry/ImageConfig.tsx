import * as React from 'react';
import Box from '@mui/material/Box';
import Dialog from '@mui/material/Dialog';
import Grid from '@mui/material/Grid';
import ListItemText from '@mui/material/ListItemText';
import List from '@mui/material/List';
import Divider from '@mui/material/Divider';
import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import IconButton from '@mui/material/IconButton';
import Typography from '@mui/material/Typography';
import Paper from '@mui/material/Paper';
import CloseIcon from '@mui/icons-material/Close';
import Slide from '@mui/material/Slide';
import { TransitionProps } from '@mui/material/transitions';
import { useGetOne, useTranslate } from 'react-admin';
import { repositoryBaseResource } from './RepositoryShow';
import { Buffer } from 'buffer';


const Transition = React.forwardRef(function Transition(
    props: TransitionProps & {
        children: React.ReactElement;
    },
    ref: React.Ref<unknown>,
) {
    return <Slide direction="up" ref={ref} {...props} />;
});


export default function ImageConfigPage({ record, isOpen, handleShowFn }: any) {
    const [open, setOpen] = React.useState(false);
    const [manifest, setManifest] = React.useState(Object);
    const translate = useTranslate();

    const { data, isLoading, error } = useGetOne(
        repositoryBaseResource,
        { id: 'blobs', meta: { name: record.repository_name, digest: record.digest } }
    );

    const decodeConfig = (data: any): any => {
        const cfg = Buffer.from(data.value, 'base64').toString('ascii');
        return JSON.parse(cfg);
    }

    React.useEffect(() => {
        setOpen(isOpen);
    }, [isOpen])


    React.useEffect(() => {
        if (isLoading || !data) {
            return
        }
        const cfg = decodeConfig(data);
        if (cfg && cfg !== null) {
            setManifest(cfg);
        }
    }, [isLoading,data])

    // if (isLoading) { return <Loading />; }

    const MainData = () => {
        return manifest && (
            <List dense={true}>
                {manifest.architecture ? <ListItemText disableTypography primary={<div style={{ fontWeight: "bolder", float: "left" }}>{"Arch: "}</div>} secondary={manifest.architecture} /> : null}
                {manifest.created ? <ListItemText disableTypography primary={<div style={{ fontWeight: "bolder", float: "left" }}>{"CreatedAt: "}</div>} secondary={manifest.created} /> : null}
                {manifest.os ? <ListItemText disableTypography primary={<div style={{ fontWeight: "bolder", float: "left" }}>{"OS: "}</div>} secondary={manifest.os} /> : null}
            </List>
        );
    }

    const ConfigData = () => {
        const  config: any  = manifest.config;
        return config && (
            <List dense={true}>
                {config.ExposedPorts ? <ListItemText disableTypography primary={<div style={{ fontWeight: "bolder", float: "left" }}>Exposed port:</div>} secondary={JSON.stringify(config.ExposedPorts)} /> : null}
                {config.Env ? <ListItemText disableTypography primary={<div style={{ fontWeight: "bolder", float: "left" }}>ENV:</div>} secondary={JSON.stringify(config.Env)} /> : null}
                {config.Cmd ? <ListItemText disableTypography primary={<div style={{ fontWeight: "bolder", float: "left" }}>CMD:</div>} secondary={JSON.stringify(config.Cmd)} /> : null}
                {config.Labels ? <ListItemText disableTypography primary={<div style={{ fontWeight: "bolder", float: "left" }}>Labels:</div>} secondary={JSON.stringify(config.Labels)} /> : null}
                {config.ArgsEscaped ? <ListItemText disableTypography primary={<div style={{ fontWeight: "bolder", float: "left" }}>ArgsEscaped:</div>} secondary={JSON.stringify(config.ArgsEscaped)} /> : null}
                {config.OnBuild ? <ListItemText disableTypography primary={<div style={{ fontWeight: "bolder", float: "left" }}>OnBuild:</div>} secondary={JSON.stringify(config.OnBuild)} /> : null}
            </List>
        );
    }

    const HistoryData = () => {
        // const config = decodeConfig(data);
        const imageHistory = manifest.history
        return imageHistory && (
            < List dense={true} >
                {imageHistory.map((item: any, index: number) => {
                    return (
                        < div key={index}>
                            {
                                item.created ? <ListItemText
                                    disableTypography
                                    primary={<div style={{ fontWeight: "bolder", float: "left" }}>
                                        Created:</div>
                                    }
                                    secondary={item.created} />
                                    : null}

                            {item.created_by ? <ListItemText
                                disableTypography
                                primary={<div style={{ fontWeight: "bolder", float: "left" }}>
                                    Created by:</div>
                                }
                                secondary={item.created_by} />
                                : null}

                            {item.comment ? <ListItemText
                                disableTypography
                                primary={<div style={{ fontWeight: "bolder", float: "left" }}>
                                    Comment:</div>
                                }
                                secondary={item.comment} />
                                : null
                            }
                            <Divider />
                        </div>
                    )

                })}
            </List >
        );
    }

    const handleClose = () => {
        handleShowFn(false);
    };


    return (
        <div>
            <Dialog
                fullScreen
                open={open}
                onClose={handleClose}

                TransitionComponent={Transition}
            >
                <AppBar sx={{ position: 'relative' }}>
                    <Toolbar>
                        <IconButton
                            edge="start"
                            color="inherit"
                            onClick={handleClose}
                            aria-label="close"
                        >
                            <CloseIcon />
                        </IconButton>
                        <Typography sx={{ ml: 2, flex: 1 }} variant="h6" component="div">
                            {record.repository_name}:{record.tag}
                        </Typography>
                    </Toolbar>
                </AppBar>
                {!isLoading && error === null ?
                    <Box sx={{ flexGrow: 1, padding: 2 }}>
                        <Grid container spacing={2}>
                            {/* ----------- MAIN SECTION ----------- */}
                            <Grid item xs={12} md={6}>
                                <Typography sx={{ mt: 4, mb: 2 }} variant="h6" component="div">
                                {translate('resources.repository.image_platform_details')}
                                </Typography>
                                <Paper elevation={3} sx={{ paddingLeft: 1 }}>
                                    {<MainData />}
                                </Paper>
                            </Grid>

                            {/* ----------- CONFIG SECTION ----------- */}

                            <Grid item xs={12} md={6}>
                                <Typography sx={{ mt: 4, mb: 2 }} variant="h6" component="div">
                                {translate('resources.repository.image_config_details')}
                                </Typography>
                                <Paper elevation={3} sx={{ paddingLeft: 1 }}>
                                    {<ConfigData />}
                                </Paper>
                            </Grid>

                            {/* ----------- HISTORY SECTION ----------- */}

                            <Grid item xs={12} md={8}>
                                <Typography sx={{ mt: 4, mb: 2 }} variant="h6" component="div">
                                {translate('resources.repository.image_history_details')}
                                </Typography>
                                <Paper elevation={3} sx={{ paddingLeft: 1 }}>
                                    {<HistoryData />}
                                </Paper>
                            </Grid>
                        </Grid>
                    </Box>
                    :
                    (
                        translate('resources.repository.message_config_data_not_loading'))}
            </Dialog>
        </div >
    );
}
