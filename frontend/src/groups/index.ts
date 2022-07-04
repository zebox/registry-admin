import GroupIcon from '@mui/icons-material/Group';
import GroupList from './GroupList';
import GroupEdit from './GroupEdit';
import GroupCreate from './GroupCreate';

const users = {
    list: GroupList,
    edit: GroupEdit,
    create:GroupCreate,
    icon: GroupIcon
};

export default users;