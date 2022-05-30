import asyncio
import datetime
import time

import sqlalchemy
import tzlocal


async def is_host_available(host: str, port: int, timeout: int = 10) -> bool:
    """Check if the specified host is reachable on the specified port

    :param host: The hostname or ip address which shall be checked
    :param port: The port which shall be checked
    :param timeout: Max. duration of the check
    :return: A boolean indicating the status
    """
    _end_time = time.time() + timeout
    while time.time() < _end_time:
        try:
            # Try to open a connection to the specified host and port and wait a maximum time of five seconds
            _s_reader, _s_writer = await asyncio.wait_for(asyncio.open_connection(host, port), timeout=5)
            # Close the stream writer again
            _s_writer.close()
            # Wait until the writer is closed
            await _s_writer.wait_closed()
            return True
        except:
            # Since the connection could not be opened wait 5 seconds before trying again
            await asyncio.sleep(5)
    return False


def get_last_schema_update(schema_name: str, engine: sqlalchemy.engine.Engine) -> datetime.datetime:
    query = (
        f"SELECT timestamp "
        f"FROM public.audit "
        f"WHERE schema_name = '{schema_name}' "
        f"ORDER BY timestamp DESC "
        f"LIMIT 1"
    )
    result = engine.execute(query).all()
    if not result:
        return datetime.datetime.now(tz=tzlocal.get_localzone())
    return datetime.datetime.fromtimestamp(result[0], tz=tzlocal.get_localzone())
