import {useRecordContext} from 'react-admin';
import DoDisturbIcon from '@mui/icons-material/DoDisturb';

export const DisabledField = ({ source }: any) => {
    const record = useRecordContext();
    return (
        <>
            {record[source] ? <DoDisturbIcon sx={{color:"red"}}/> : ""}
        </>
    )
}