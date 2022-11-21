import Box from '@mui/material/Box';
import {
    useTranslate,
    MenuItemLink,
    MenuProps,
    useSidebarState,
    usePermissions,
    NullableBooleanInputClasses
} from 'react-admin';

import users from '../users';
import groups from '../groups';
import access from '../access';
import repository from '../registry';

import {requirePermission} from '../helpers/Helpers';

const Menu = ({dense = false}: MenuProps) => {

    const {permissions} = usePermissions();
    const translate = useTranslate();
    const [open] = useSidebarState();

    return (<Box
            sx={{
                width: open ? 200 : 50,
                marginTop: 1,
                marginBottom: 1,
                transition: theme =>
                    theme.transitions.create('width', {
                        easing: theme.transitions.easing.sharp,
                        duration: theme.transitions.duration.leavingScreen,
                    }),
            }}
        >
            <MenuItemLink
                to="/registry/catalog"
                state={{_scrollToTop: true}}
                primaryText={translate(`resources.commands.repository_name`, {
                    smart_count: 2,
                })}
                leftIcon={<repository.icon/>}
                dense={dense}
            />
            {requirePermission(permissions, 'admin') || requirePermission(permissions, 'manager') ?
                <MenuItemLink
                    to="/access"
                    state={{_scrollToTop: true}}
                    primaryText={translate(`resources.commands.access_name`, {
                        smart_count: 2,
                    })}
                    leftIcon={<access.icon/>}
                    dense={dense}
                /> : null
            }
            {requirePermission(permissions, 'admin') ? <>
                <MenuItemLink
                    to="/users"
                    state={{_scrollToTop: true}}
                    primaryText={translate(`resources.commands.users_name`, {
                        smart_count: 2,
                    })}
                    leftIcon={<users.icon/>}
                    dense={dense}
                />
                <MenuItemLink
                    to="/groups"
                    state={{_scrollToTop: true}}
                    primaryText={translate(`resources.commands.groups_name`, {
                        smart_count: 2,
                    })}
                    leftIcon={<groups.icon/>}
                    dense={dense}
                />
            </> : null}
    </Box>


    );
};

export default Menu;