import * as React from "react";
import Tooltip from '@mui/material/Tooltip';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import { useTranslate } from "react-admin";
import IconButton from '@mui/material/IconButton';

const CopyToClipboard = (props: any): React.ReactElement => {
  const [showTooltip, setShowTooltip] = React.useState<boolean>(false);
  const translate=useTranslate();

  const onCopy = (content: any): void => {
    navigator.clipboard.writeText(content);
    setShowTooltip(true);
  }

  return (
    <Tooltip
      open={showTooltip}
      title={translate('resources.repository.message_copied_to_clipboard')}
      leaveDelay={1500}
      onClose={() => { setShowTooltip(false) }}
    >
      <IconButton onClick={() => onCopy(props.content ? props.content : '')}>
        <ContentCopyIcon  />
      </IconButton>

    </Tooltip>
  );
}



export default CopyToClipboard;
